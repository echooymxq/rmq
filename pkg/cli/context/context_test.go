package contextcmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/echooymxq/rmq/pkg/config"
	"github.com/spf13/cobra"
)

func TestValidateAddContextInputReportsRequiredParams(t *testing.T) {
	cmd := &cobra.Command{Use: "add"}

	_, _, err := validateAddContextInput(cmd, nil, "")
	if err == nil {
		t.Fatal("expected required parameter error")
	}

	message := err.Error()
	assertContains(t, message, "missing required parameter(s): NAME; -n, --nameserver")
	assertContains(t, message, "Required:")
	assertContains(t, message, "Usage:")
	assertContains(t, message, "Example:")
}

func TestValidateAddContextInputAcceptsNameAndNameserver(t *testing.T) {
	cmd := &cobra.Command{Use: "add"}

	name, addrs, err := validateAddContextInput(cmd, []string{" prod "}, "10.0.0.1:9876, 10.0.0.2:9876")
	if err != nil {
		t.Fatal(err)
	}

	if name != "prod" {
		t.Fatalf("got name %q, want %q", name, "prod")
	}
	if len(addrs) != 2 {
		t.Fatalf("got %d nameserver addresses, want 2", len(addrs))
	}
	if addrs[0] != "10.0.0.1:9876" {
		t.Fatalf("got first nameserver %q, want %q", addrs[0], "10.0.0.1:9876")
	}
	if addrs[1] != "10.0.0.2:9876" {
		t.Fatalf("got second nameserver %q, want %q", addrs[1], "10.0.0.2:9876")
	}
}

func TestValidateUseContextInputReportsMissingName(t *testing.T) {
	cmd := &cobra.Command{Use: "use"}

	_, err := validateUseContextInput(cmd, nil)
	if err == nil {
		t.Fatal("expected required parameter error")
	}

	message := err.Error()
	assertContains(t, message, "missing required parameter(s): NAME")
	assertContains(t, message, "Required:")
	assertContains(t, message, "Usage:")
	assertContains(t, message, "Example:")
}

func TestValidateUseContextInputRejectsExtraArgs(t *testing.T) {
	cmd := &cobra.Command{Use: "use"}

	_, err := validateUseContextInput(cmd, []string{"prod", "extra"})
	if err == nil {
		t.Fatal("expected extra argument error")
	}

	assertContains(t, err.Error(), "too many arguments: extra")
}

func TestValidateUseContextInputAcceptsName(t *testing.T) {
	cmd := &cobra.Command{Use: "use"}

	name, err := validateUseContextInput(cmd, []string{" prod "})
	if err != nil {
		t.Fatal(err)
	}
	if name != "prod" {
		t.Fatalf("got name %q, want %q", name, "prod")
	}
}

func TestValidateNoContextArgsRejectsExtraArgs(t *testing.T) {
	cmd := &cobra.Command{Use: "list"}

	err := validateNoContextArgs(cmd, []string{"extra"})
	if err == nil {
		t.Fatal("expected extra argument error")
	}

	assertContains(t, err.Error(), "unexpected argument(s): extra")
	assertContains(t, err.Error(), "Usage:")
}

func TestAddCredentialFlagsHaveShortFlags(t *testing.T) {
	cmd := Add(new(config.RocketMQConfig))

	accessKey := cmd.Flags().Lookup("accessKey")
	if accessKey == nil {
		t.Fatal("expected accessKey flag")
	}
	if accessKey.Shorthand != "a" {
		t.Fatalf("got accessKey shorthand %q, want %q", accessKey.Shorthand, "a")
	}

	secretKey := cmd.Flags().Lookup("secretKey")
	if secretKey == nil {
		t.Fatal("expected secretKey flag")
	}
	if secretKey.Shorthand != "s" {
		t.Fatalf("got secretKey shorthand %q, want %q", secretKey.Shorthand, "s")
	}
}

func TestAddRejectsDuplicateBeforeCredentialPrompt(t *testing.T) {
	configFile := filepath.Join(t.TempDir(), "rmq.yaml")
	if err := os.WriteFile(configFile, []byte(`
current: local
contexts:
  local:
    namesrvAddrs:
      - localhost:9876
`), 0600); err != nil {
		t.Fatal(err)
	}

	cmd := Add(&config.RocketMQConfig{ConfigFile: configFile})
	var stderr bytes.Buffer
	cmd.SetArgs([]string{"local", "-n", "localhost:9876"})
	cmd.SetErr(&stderr)
	cmd.SetIn(strings.NewReader("y\n"))

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected duplicate context error")
	}
	if err.Error() != `context "local" already exists` {
		t.Fatalf("got error %q, want duplicate context error", err.Error())
	}
	if strings.Contains(stderr.String(), "Missing credential") {
		t.Fatalf("got credential prompt %q, want no prompt", stderr.String())
	}
}

func TestConfirmMissingCredentialsSkipsPromptWhenCredentialsComplete(t *testing.T) {
	cmd := &cobra.Command{Use: "add"}
	var stderr bytes.Buffer
	cmd.SetErr(&stderr)

	if err := confirmMissingCredentials(cmd, "ak", "sk"); err != nil {
		t.Fatal(err)
	}
	if stderr.Len() != 0 {
		t.Fatalf("got stderr %q, want empty", stderr.String())
	}
}

func TestConfirmMissingCredentialsAllowsYes(t *testing.T) {
	cmd := &cobra.Command{Use: "add"}
	var stderr bytes.Buffer
	cmd.SetIn(strings.NewReader("y\n"))
	cmd.SetErr(&stderr)

	if err := confirmMissingCredentials(cmd, "", ""); err != nil {
		t.Fatal(err)
	}
	assertContains(t, stderr.String(), "Missing credential(s): -a, --accessKey, -s, --secretKey.")
	assertContains(t, stderr.String(), "Continue without complete credentials? [y/N]")
}

func TestConfirmMissingCredentialsRejectsDefaultAnswer(t *testing.T) {
	cmd := &cobra.Command{Use: "add"}
	var stderr bytes.Buffer
	cmd.SetIn(strings.NewReader("\n"))
	cmd.SetErr(&stderr)

	if err := confirmMissingCredentials(cmd, "", ""); err == nil {
		t.Fatal("expected confirmation rejection")
	}
}

func assertContains(t *testing.T, value, want string) {
	t.Helper()
	if !strings.Contains(value, want) {
		t.Fatalf("got %q, want it to contain %q", value, want)
	}
}
