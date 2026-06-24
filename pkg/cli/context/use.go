package contextcmd

import (
	"fmt"

	"github.com/echooymxq/rmq/pkg/config"
	"github.com/spf13/cobra"
)

func Use(r *config.RocketMQConfig) *cobra.Command {
	return &cobra.Command{
		Use:   "use NAME",
		Short: "Set the current configuration context",
		RunE: func(cmd *cobra.Command, args []string) error {
			name, err := validateUseContextInput(cmd, args)
			if err != nil {
				return err
			}
			if err := r.SetCurrentContext(name); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Switched to context %q.\n", name)
			return nil
		},
	}
}
