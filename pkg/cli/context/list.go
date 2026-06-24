package contextcmd

import (
	"fmt"
	"strings"

	"github.com/echooymxq/rmq/pkg/config"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

func List(r *config.RocketMQConfig) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List configuration contexts",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateNoContextArgs(cmd, args); err != nil {
				return err
			}
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
