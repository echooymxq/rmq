package contextcmd

import (
	"fmt"

	"github.com/echooymxq/rmq/pkg/config"
	"github.com/spf13/cobra"
)

func Current(r *config.RocketMQConfig) *cobra.Command {
	return &cobra.Command{
		Use:   "current",
		Short: "Print the current configuration context",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateNoContextArgs(cmd, args); err != nil {
				return err
			}
			store, err := r.LoadContextStore()
			if err != nil {
				return err
			}
			if store.Current == "" {
				return fmt.Errorf("current context is not set")
			}
			fmt.Fprintln(cmd.OutOrStdout(), store.Current)
			return nil
		},
	}
}
