package config

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
)

func TestResolveUsesCurrentContext(t *testing.T) {
	configFile := writeConfig(t, `
current: prod
contexts:
  prod:
    namesrvAddrs:
      - 10.0.0.1:9876
    accessKey: prod-ak
    secretKey: prod-sk
    timeout: 5s
`)
	r := executeConfig(t, "--config", configFile)

	assertEqual(t, r.Runtime.Context, "prod")
	assertEqual(t, r.Runtime.NamesrvAddrs[0], "10.0.0.1:9876")
	assertEqual(t, r.Runtime.AccessKey, "prod-ak")
	assertEqual(t, r.Runtime.SecretKey, "prod-sk")
	assertEqual(t, r.Runtime.Timeout.String(), "5s")
}

func TestResolveAppliesFlagOverrides(t *testing.T) {
	configFile := writeConfig(t, `
current: prod
contexts:
  prod:
    namesrvAddrs:
      - 10.0.0.1:9876
    accessKey: prod-ak
    secretKey: prod-sk
    timeout: 5s
  staging:
    namesrvAddrs:
      - 10.0.0.2:9876
    accessKey: staging-ak
    secretKey: staging-sk
`)
	r := executeConfig(t,
		"--config", configFile,
		"--context", "staging",
		"-n", "127.0.0.1:9876,127.0.0.2:9876",
		"--accessKey", "flag-ak",
		"--secretKey", "flag-sk",
		"--timeout", "2s",
	)

	assertEqual(t, r.Runtime.Context, "staging")
	assertEqual(t, r.Runtime.NamesrvAddrs[0], "127.0.0.1:9876")
	assertEqual(t, r.Runtime.NamesrvAddrs[1], "127.0.0.2:9876")
	assertEqual(t, r.Runtime.AccessKey, "flag-ak")
	assertEqual(t, r.Runtime.SecretKey, "flag-sk")
	assertEqual(t, r.Runtime.Timeout.String(), "2s")
}

func TestResolveRejectsTopLevelLegacyConfig(t *testing.T) {
	configFile := writeConfig(t, `
AccessKey: legacy-ak
SecretKey: legacy-sk
NamesrvAddrs: 192.168.0.1:9876,192.168.0.2:9876
`)
	_, err := runConfig("--config", configFile)
	if err == nil {
		t.Fatal("expected top-level config error")
	}
}

func TestResolveRequiresContextWhenConfigHasNoCurrent(t *testing.T) {
	configFile := writeConfig(t, `
contexts:
  prod:
    namesrvAddrs:
      - 10.0.0.1:9876
`)
	_, err := runConfig("--config", configFile)
	if err == nil {
		t.Fatal("expected missing current context error")
	}
}

func TestResolveConfigFileUsesDefaultConfigPath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}

	r := new(RocketMQConfig)
	configFile, explicit, err := r.resolveConfigFile()
	if err != nil {
		t.Fatal(err)
	}

	assertEqual(t, configFile, filepath.Join(home, ".config", "rmq.yaml"))
	if explicit {
		t.Fatal("expected default config path to be implicit")
	}
}

func TestLoadContextStore(t *testing.T) {
	configFile := writeConfig(t, `
current: prod
contexts:
  staging:
    namesrvAddrs:
      - 10.0.0.2:9876
    timeout: 2s
  prod:
    namesrvAddrs:
      - 10.0.0.1:9876
    accessKey: prod-ak
    secretKey: prod-sk
    timeout: 5s
`)
	r := &RocketMQConfig{ConfigFile: configFile}

	store, err := r.LoadContextStore()
	if err != nil {
		t.Fatal(err)
	}

	assertEqual(t, store.ConfigFile, configFile)
	assertEqual(t, store.Current, "prod")
	assertEqual(t, store.Contexts[0].Name, "prod")
	assertEqual(t, store.Contexts[0].NamesrvAddrs[0], "10.0.0.1:9876")
	assertEqual(t, store.Contexts[0].AccessKey, "prod-ak")
	assertEqual(t, store.Contexts[0].Timeout, "5s")
	if !store.Contexts[0].Current {
		t.Fatal("expected prod to be current")
	}
	assertEqual(t, store.Contexts[1].Name, "staging")
}

func TestLoadContextStoreAllowsMissingDefaultConfig(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	r := new(RocketMQConfig)

	store, err := r.LoadContextStore()
	if err != nil {
		t.Fatal(err)
	}
	assertEqual(t, store.ConfigFile, filepath.Join(home, ".config", "rmq.yaml"))
	if len(store.Contexts) != 0 {
		t.Fatalf("got %d contexts, want 0", len(store.Contexts))
	}
}

func TestLoadContextStoreRejectsMissingExplicitConfig(t *testing.T) {
	r := &RocketMQConfig{ConfigFile: filepath.Join(t.TempDir(), "missing.yaml")}

	if _, err := r.LoadContextStore(); err == nil {
		t.Fatal("expected missing explicit config error")
	}
}

func TestSetCurrentContext(t *testing.T) {
	configFile := writeConfig(t, `
current: prod
contexts:
  prod:
    namesrvAddrs:
      - 10.0.0.1:9876
  staging:
    namesrvAddrs:
      - 10.0.0.2:9876
`)
	r := &RocketMQConfig{ConfigFile: configFile}

	if err := r.SetCurrentContext("staging"); err != nil {
		t.Fatal(err)
	}

	cfg, _, err := readConfigFile(configFile, true)
	if err != nil {
		t.Fatal(err)
	}
	assertEqual(t, cfg.Current, "staging")
	if _, ok := cfg.Contexts["prod"]; !ok {
		t.Fatal("expected prod context to remain")
	}
}

func TestSetCurrentContextRejectsMissingDefaultConfig(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	r := new(RocketMQConfig)

	if err := r.SetCurrentContext("prod"); err == nil {
		t.Fatal("expected missing config error")
	}
}

func TestSetCurrentContextRejectsUnknownContext(t *testing.T) {
	configFile := writeConfig(t, `
current: prod
contexts:
  prod:
    namesrvAddrs:
      - 10.0.0.1:9876
`)
	r := &RocketMQConfig{ConfigFile: configFile}

	if err := r.SetCurrentContext("missing"); err == nil {
		t.Fatal("expected missing context error")
	}
}

func TestShouldSkipResolveChecksParents(t *testing.T) {
	root := &cobra.Command{Use: "rmq"}
	parent := &cobra.Command{
		Use: "context",
		Annotations: map[string]string{
			AnnotationSkipResolve: "true",
		},
	}
	child := &cobra.Command{Use: "current"}
	parent.AddCommand(child)
	root.AddCommand(parent)

	if !ShouldSkipResolve(child) {
		t.Fatal("expected child command to skip resolve")
	}
	if ShouldSkipResolve(root) {
		t.Fatal("expected root command to resolve")
	}
}

func executeConfig(t *testing.T, args ...string) *RocketMQConfig {
	t.Helper()
	r, err := runConfig(args...)
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func runConfig(args ...string) (*RocketMQConfig, error) {
	r := new(RocketMQConfig)
	cmd := &cobra.Command{
		Use: "rmq",
		RunE: func(cmd *cobra.Command, args []string) error {
			return r.Resolve(cmd)
		},
	}
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs(args)
	r.InstallRootFlags(cmd)
	return r, cmd.Execute()
}

func writeConfig(t *testing.T, data string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "rmq.yaml")
	if err := os.WriteFile(path, []byte(data), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

func assertEqual(t *testing.T, got, want string) {
	t.Helper()
	if got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
}
