package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func MarkFlagsRequired(cmd *cobra.Command, names ...string) {
	previous := cmd.PreRunE
	cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		for _, name := range names {
			flag := cmd.Flags().Lookup(name)
			if flag == nil {
				panic(fmt.Sprintf("required flag %q is not defined", name))
			}
			if flag.Changed {
				continue
			}
			return fmt.Errorf("required flag %q not set, use %s", name, flagUsage(flag.Shorthand, flag.Name))
		}
		if previous != nil {
			return previous(cmd, args)
		}
		return nil
	}
}

func flagUsage(shorthand, name string) string {
	var forms []string
	if shorthand != "" {
		forms = append(forms, "-"+shorthand)
	}
	forms = append(forms, "--"+name)
	return strings.Join(forms, ", ")
}

func ChainPreRunE(cmd *cobra.Command, fn func(*cobra.Command, []string) error) {
	previous := cmd.PreRunE
	cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		if previous != nil {
			if err := previous(cmd, args); err != nil {
				return err
			}
		}
		return fn(cmd, args)
	}
}
