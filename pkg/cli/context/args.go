package contextcmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func validateNoContextArgs(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return nil
	}
	return fmt.Errorf("unexpected argument(s): %s\n\nUsage:\n  %s", strings.Join(args, " "), cmd.CommandPath())
}

func validateUseContextInput(cmd *cobra.Command, args []string) (string, error) {
	if len(args) > 1 {
		return "", fmt.Errorf("too many arguments: %s\n\n%s", strings.Join(args[1:], " "), useContextRequiredMessage(cmd))
	}

	name := ""
	if len(args) == 1 {
		name = strings.TrimSpace(args[0])
	}
	if name == "" {
		return "", fmt.Errorf("missing required parameter(s): NAME\n\n%s", useContextRequiredMessage(cmd))
	}

	return name, nil
}

func useContextRequiredMessage(cmd *cobra.Command) string {
	commandPath := cmd.CommandPath()
	return fmt.Sprintf(`Required:
  NAME  context name, such as prod

Usage:
  %s NAME

Example:
  %s prod`, commandPath, commandPath)
}
