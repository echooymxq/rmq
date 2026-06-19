package contextcmd

import (
	"fmt"
	"strings"

	"github.com/echooymxq/rmq/pkg/config"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

func NewCommand(r *config.RocketMQConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "context",
		Short: "Manage configuration contexts",
		Annotations: map[string]string{
			config.AnnotationSkipResolve: "true",
		},
	}
	cmd.AddCommand(List(r))
	cmd.AddCommand(Current(r))
	cmd.AddCommand(Use(r))
	return cmd
}

func List(r *config.RocketMQConfig) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List configuration contexts",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := r.LoadContextStore()
			if err != nil {
				return err
			}
			if len(store.Contexts) == 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "No contexts configured. Create %s or pass --config.\n", store.ConfigFile)
				return nil
			}

			table := tablewriter.NewWriter(cmd.OutOrStdout())
			table.SetHeader([]string{"Current", "Name", "Nameservers"})
			for _, ctx := range store.Contexts {
				current := ""
				if ctx.Current {
					current = "*"
				}
				table.Append([]string{
					current,
					ctx.Name,
					strings.Join(ctx.NamesrvAddrs, ","),
				})
			}
			table.Render()
			return nil
		},
	}
}

func Current(r *config.RocketMQConfig) *cobra.Command {
	return &cobra.Command{
		Use:   "current",
		Short: "Print the current configuration context",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
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

func Use(r *config.RocketMQConfig) *cobra.Command {
	return &cobra.Command{
		Use:   "use NAME",
		Short: "Set the current configuration context",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := strings.TrimSpace(args[0])
			if err := r.SetCurrentContext(name); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Switched to context %q.\n", name)
			return nil
		},
	}
}
