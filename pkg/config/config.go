package config

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

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
)

const AnnotationSkipResolve = "rmq.echooymxq.io/skip-resolve"

type RocketMQConfig struct {
	ConfigFile   string
	Context      string
	AccessKey    string
	SecretKey    string
	NamesrvAddrs string

	Runtime RuntimeConfig
}

type RuntimeConfig struct {
	Context      string
	ConfigFile   string
	NamesrvAddrs []string
	AccessKey    string
	SecretKey    string
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
}

type fileConfig struct {
	Current  string                   `yaml:"current"`
	Contexts map[string]contextConfig `yaml:"contexts"`
}

type contextConfig struct {
	NamesrvAddrs stringList `yaml:"namesrvAddrs"`
	AccessKey    string     `yaml:"accessKey,omitempty"`
	SecretKey    string     `yaml:"secretKey,omitempty"`
}

type stringList []string

func (s *stringList) UnmarshalYAML(value *yaml.Node) error {
	switch value.Kind {
	case yaml.ScalarNode:
		if value.Value == "" {
			*s = nil
			return nil
		}
		*s = SplitAndTrim(value.Value)
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

	r.Runtime = runtime
	r.NamesrvAddrs = strings.Join(runtime.NamesrvAddrs, ",")
	r.AccessKey = runtime.AccessKey
	r.SecretKey = runtime.SecretKey
	return nil
}

func (r *RocketMQConfig) GetNamesrvAddrs() []string {
	if len(r.Runtime.NamesrvAddrs) > 0 {
		return append([]string(nil), r.Runtime.NamesrvAddrs...)
	}
	return SplitAndTrim(r.NamesrvAddrs)
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
		})
	}
	return store, nil
}

func (r *RocketMQConfig) ContextExists(name string) (bool, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return false, errors.New("context name is required")
	}

	configFile, explicitConfig, err := r.resolveConfigFile()
	if err != nil {
		return false, err
	}
	cfg, loaded, err := readConfigFile(configFile, explicitConfig)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	if !loaded || cfg.Contexts == nil {
		return false, nil
	}

	_, ok := cfg.Contexts[name]
	return ok, nil
}

func (r *RocketMQConfig) AddContext(name string, namesrvAddrs []string, accessKey, secretKey string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.New("context name is required")
	}
	if len(namesrvAddrs) == 0 {
		return errors.New("nameserver is required")
	}

	configFile, explicitConfig, err := r.resolveConfigFile()
	if err != nil {
		return err
	}
	cfg, loaded, err := readConfigFile(configFile, explicitConfig)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	}
	if !loaded {
		cfg = fileConfig{
			Contexts: make(map[string]contextConfig),
		}
	}
	if cfg.Contexts == nil {
		cfg.Contexts = make(map[string]contextConfig)
	}
	if _, ok := cfg.Contexts[name]; ok {
		return fmt.Errorf("context %q already exists", name)
	}

	cfg.Contexts[name] = contextConfig{
		NamesrvAddrs: stringList(namesrvAddrs),
		AccessKey:    accessKey,
		SecretKey:    secretKey,
	}
	if cfg.Current == "" {
		cfg.Current = name
	}

	return writeConfigFile(configFile, cfg)
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

	return writeConfigFile(configFile, cfg)
}

func writeConfigFile(configFile string, cfg fileConfig) error {
	mode := os.FileMode(0600)
	if info, err := os.Stat(configFile); err == nil {
		mode = info.Mode().Perm()
	}

	if err := os.MkdirAll(filepath.Dir(configFile), 0700); err != nil {
		return err
	}

	var data bytes.Buffer
	encoder := yaml.NewEncoder(&data)
	encoder.SetIndent(2)
	if err := encoder.Encode(&cfg); err != nil {
		_ = encoder.Close()
		return err
	}
	if err := encoder.Close(); err != nil {
		return err
	}

	return os.WriteFile(configFile, data.Bytes(), mode)
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
	return nil
}

func applyFlagOverrides(cmd *cobra.Command, r *RocketMQConfig, runtime *RuntimeConfig) error {
	if cmd == nil {
		if r.NamesrvAddrs != "" {
			runtime.NamesrvAddrs = SplitAndTrim(r.NamesrvAddrs)
		}
		if r.AccessKey != "" {
			runtime.AccessKey = r.AccessKey
		}
		if r.SecretKey != "" {
			runtime.SecretKey = r.SecretKey
		}
		return nil
	}
	if flagChanged(cmd, FlagNameserver) {
		runtime.NamesrvAddrs = SplitAndTrim(r.NamesrvAddrs)
	}
	if flagChanged(cmd, FlagAccessKey) {
		runtime.AccessKey = r.AccessKey
	}
	if flagChanged(cmd, FlagSecretKey) {
		runtime.SecretKey = r.SecretKey
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

func SplitAndTrim(value string) []string {
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
