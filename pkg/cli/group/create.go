package group

import (
	"context"
	"fmt"

	"github.com/apache/rocketmq-client-go/v2/admin"
	"github.com/echooymxq/rmq/pkg/cli"
	"github.com/echooymxq/rmq/pkg/config"
	"github.com/echooymxq/rmq/pkg/rocketmq"
	"github.com/spf13/cobra"
)

func Create(r *config.RocketMQConfig) *cobra.Command {
	var (
		group string
	)
	var cmd = &cobra.Command{
		Use: "create",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := rocketmq.NewAdminClient(r)
			if err != nil {
				return err
			}
			defer rocketmq.Close(client)

			err = client.CreateSubscriptionGroup(context.Background(),
				admin.WithGroupName(group),
			)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Create consumer group %q success.\n", group)
			return nil
		},
	}

	cmd.Flags().StringVarP(&group, "group", "g", "", "")
	cli.MarkFlagsRequired(cmd, "group")
	return cmd
}
