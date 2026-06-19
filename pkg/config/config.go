package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

const (
	DefaultNamesrvAddr = "127.0.0.1:9876"
	FlagAccessKey      = "accessKey"
	FlagConfig         = "config"
	FlagContext        = "context"
	FlagNameserver     = "nameserver"
	FlagSecretKey      = "secretKey"
	FlagTimeout        = "timeout"
)

const AnnotationSkipResolve = "rmq.echooymxq.io/skip-resolve"

type RocketMQConfig struct {
	ConfigFile   string
	Context      string
	AccessKey    string
	SecretKey    string
	NamesrvAddrs string
	Timeout      string

	Runtime RuntimeConfig
}

type RuntimeConfig struct {
	Context      string
	ConfigFile   string
	NamesrvAddrs []string
	AccessKey    string
	SecretKey    string
	Timeout      time.Duration
}

type ContextStore struct {
	ConfigFile string
	Current    string
	Contexts   []ContextSummary
}

type ContextSummary struct {
	Name         string
	Current      bool
	NamesrvAddrs []string
	AccessKey    string
	Timeout      string
}

type fileConfig struct {
	Current  string                   `yaml:"current"`
	Contexts map[string]contextConfig `yaml:"contexts"`
}

type contextConfig struct {
	NamesrvAddrs stringList `yaml:"namesrvAddrs"`
	AccessKey    string     `yaml:"accessKey"`
	SecretKey    string     `yaml:"secretKey"`
	Timeout      string     `yaml:"timeout"`
}

type stringList []string

func (s *stringList) UnmarshalYAML(value *yaml.Node) error {
	switch value.Kind {
	case yaml.ScalarNode:
		if value.Value == "" {
			*s = nil
			return nil
		}
		*s = splitAndTrim(value.Value)
		return nil
	case yaml.SequenceNode:
		var values []string
		for _, node := range value.Content {
			if node.Kind != yaml.ScalarNode {
				return fmt.Errorf("expected string value in list")
			}
			values = append(values, node.Value)
		}
		*s = cleanList(values)
		return nil
	default:
		return fmt.Errorf("expected string or list")
	}
}

func (r *RocketMQConfig) Load() error {
	return r.Resolve(nil)
}

func (r *RocketMQConfig) InstallRootFlags(cmd *cobra.Command) {
	pf := cmd.PersistentFlags()
	pf.StringVar(&r.ConfigFile, FlagConfig, "", "config file")
	pf.StringVar(&r.Context, FlagContext, "", "configuration context")
	pf.StringVarP(&r.NamesrvAddrs, FlagNameserver, "n", "", "comma separated list of namesrv host:ports")
	pf.StringVar(&r.AccessKey, FlagAccessKey, "", "access key")
	pf.StringVar(&r.SecretKey, FlagSecretKey, "", "secret key")
	pf.StringVar(&r.Timeout, FlagTimeout, "", "request timeout, such as 3s or 500ms")
}

func (r *RocketMQConfig) InstallRocketMQFlags(cmd *cobra.Command) {
	// Connection flags are installed on the root command and inherited by all subcommands.
}

func ShouldSkipResolve(cmd *cobra.Command) bool {
	for current := cmd; current != nil; current = current.Parent() {
		if current.Annotations[AnnotationSkipResolve] == "true" {
			return true
		}
	}
	return false
}

func (r *RocketMQConfig) Resolve(cmd *cobra.Command) error {
	configFile, explicitConfig, err := r.resolveConfigFile()
	if err != nil {
		return err
	}

	runtime := RuntimeConfig{
		ConfigFile:   configFile,
		NamesrvAddrs: []string{DefaultNamesrvAddr},
	}

	cfg, loaded, err := readConfigFile(configFile, explicitConfig)
	if err != nil {
		return err
	}
	if loaded {
		selected, contextName, err := selectContext(cfg, r.Context)
		if err != nil {
			return err
		}
		if err := applyContext(&runtime, selected); err != nil {
			return err
		}
		runtime.Context = contextName
	} else if r.Context != "" {
		return fmt.Errorf("context %q not found: config file %s was not loaded", r.Context, configFile)
	}

	if err := applyFlagOverrides(cmd, r, &runtime); err != nil {
		return err
	}

	if len(runtime.NamesrvAddrs) == 0 {
		return errors.New("nameserver is required")
	}
	if r.Timeout != "" || runtime.Timeout > 0 {
		if runtime.Timeout < 0 {
			return errors.New("timeout must be positive")
		}
	}

	r.Runtime = runtime
	r.NamesrvAddrs = strings.Join(runtime.NamesrvAddrs, ",")
	r.AccessKey = runtime.AccessKey
	r.SecretKey = runtime.SecretKey
	if runtime.Timeout > 0 {
		r.Timeout = runtime.Timeout.String()
	}
	return nil
}

func (r *RocketMQConfig) GetNamesrvAddrs() []string {
	if len(r.Runtime.NamesrvAddrs) > 0 {
		return append([]string(nil), r.Runtime.NamesrvAddrs...)
	}
	return splitAndTrim(r.NamesrvAddrs)
}

func (r *RocketMQConfig) LoadContextStore() (ContextStore, error) {
	configFile, explicitConfig, err := r.resolveConfigFile()
	if err != nil {
		return ContextStore{}, err
	}
	store := ContextStore{
		ConfigFile: configFile,
	}

	cfg, loaded, err := readConfigFile(configFile, explicitConfig)
	if err != nil {
		return ContextStore{}, err
	}
	if !loaded {
		return store, nil
	}
	if len(cfg.Contexts) == 0 {
		return ContextStore{}, errors.New("config file must define contexts")
	}

	names := make([]string, 0, len(cfg.Contexts))
	for name := range cfg.Contexts {
		names = append(names, name)
	}
	sort.Strings(names)

	store.Current = cfg.Current
	store.Contexts = make([]ContextSummary, 0, len(names))
	for _, name := range names {
		ctx := cfg.Contexts[name]
		store.Contexts = append(store.Contexts, ContextSummary{
			Name:         name,
			Current:      name == cfg.Current,
			NamesrvAddrs: append([]string(nil), ctx.NamesrvAddrs...),
			AccessKey:    ctx.AccessKey,
			Timeout:      ctx.Timeout,
		})
	}
	return store, nil
}

func (r *RocketMQConfig) SetCurrentContext(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.New("context name is required")
	}

	configFile, explicitConfig, err := r.resolveConfigFile()
	if err != nil {
		return err
	}
	cfg, loaded, err := readConfigFile(configFile, explicitConfig)
	if err != nil {
		return err
	}
	if !loaded {
		return fmt.Errorf("config file %s was not found", configFile)
	}
	if len(cfg.Contexts) == 0 {
		return errors.New("config file must define contexts")
	}
	if _, ok := cfg.Contexts[name]; !ok {
		return fmt.Errorf("context %q not found", name)
	}
	cfg.Current = name

	mode := os.FileMode(0600)
	if info, err := os.Stat(configFile); err == nil {
		mode = info.Mode().Perm()
	}
	data, err := yaml.Marshal(&cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(configFile, data, mode)
}

func (r *RocketMQConfig) resolveConfigFile() (string, bool, error) {
	if r.ConfigFile != "" {
		return r.ConfigFile, true, nil
	}
	configDir, err := os.UserHomeDir()
	if err != nil {
		return "", false, err
	}
	return filepath.Join(configDir, ".config", "rmq.yaml"), false, nil
}

func readConfigFile(configFile string, explicitConfig bool) (fileConfig, bool, error) {
	var cfg fileConfig
	data, err := os.ReadFile(configFile)
	if err != nil {
		if os.IsNotExist(err) && !explicitConfig {
			return cfg, false, nil
		}
		return cfg, false, err
	}
	if len(data) == 0 {
		return cfg, false, errors.New("empty config file")
	}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return cfg, false, err
	}
	return cfg, true, nil
}

func selectContext(cfg fileConfig, requested string) (contextConfig, string, error) {
	if requested != "" {
		selected, ok := cfg.Contexts[requested]
		if !ok {
			return contextConfig{}, "", fmt.Errorf("context %q not found", requested)
		}
		return selected, requested, nil
	}
	if cfg.Current != "" {
		selected, ok := cfg.Contexts[cfg.Current]
		if !ok {
			return contextConfig{}, "", fmt.Errorf("current context %q not found", cfg.Current)
		}
		return selected, cfg.Current, nil
	}
	if len(cfg.Contexts) > 0 {
		return contextConfig{}, "", errors.New("context is required when config has contexts but no current context")
	}
	return contextConfig{}, "", errors.New("config file must define contexts")
}

func applyContext(runtime *RuntimeConfig, ctx contextConfig) error {
	if len(ctx.NamesrvAddrs) > 0 {
		runtime.NamesrvAddrs = append([]string(nil), ctx.NamesrvAddrs...)
	}
	runtime.AccessKey = ctx.AccessKey
	runtime.SecretKey = ctx.SecretKey
	if ctx.Timeout != "" {
		timeout, err := time.ParseDuration(ctx.Timeout)
		if err != nil {
			return fmt.Errorf("invalid timeout %q: %w", ctx.Timeout, err)
		}
		runtime.Timeout = timeout
	}
	return nil
}

func applyFlagOverrides(cmd *cobra.Command, r *RocketMQConfig, runtime *RuntimeConfig) error {
	if cmd == nil {
		if r.NamesrvAddrs != "" {
			runtime.NamesrvAddrs = splitAndTrim(r.NamesrvAddrs)
		}
		if r.AccessKey != "" {
			runtime.AccessKey = r.AccessKey
		}
		if r.SecretKey != "" {
			runtime.SecretKey = r.SecretKey
		}
		if r.Timeout != "" {
			timeout, err := time.ParseDuration(r.Timeout)
			if err != nil {
				return fmt.Errorf("invalid timeout %q: %w", r.Timeout, err)
			}
			runtime.Timeout = timeout
		}
		return nil
	}
	if flagChanged(cmd, FlagNameserver) {
		runtime.NamesrvAddrs = splitAndTrim(r.NamesrvAddrs)
	}
	if flagChanged(cmd, FlagAccessKey) {
		runtime.AccessKey = r.AccessKey
	}
	if flagChanged(cmd, FlagSecretKey) {
		runtime.SecretKey = r.SecretKey
	}
	if flagChanged(cmd, FlagTimeout) && r.Timeout != "" {
		timeout, err := time.ParseDuration(r.Timeout)
		if err != nil {
			return fmt.Errorf("invalid timeout %q: %w", r.Timeout, err)
		}
		runtime.Timeout = timeout
	}
	return nil
}

func flagChanged(cmd *cobra.Command, name string) bool {
	if cmd == nil {
		return false
	}
	flag := cmd.Flags().Lookup(name)
	if flag == nil {
		flag = cmd.InheritedFlags().Lookup(name)
	}
	return flag != nil && flag.Changed
}

func splitAndTrim(value string) []string {
	if value == "" {
		return nil
	}
	return cleanList(strings.Split(value, ","))
}

func cleanList(values []string) []string {
	var result []string
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
