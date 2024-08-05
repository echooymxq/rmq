package group

import (
	"context"
	"github.com/apache/rocketmq-client-go/v2/admin"
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
		Run: func(cmd *cobra.Command, args []string) {
			client, _ := rocketmq.NewAdminClient(r)
			defer rocketmq.Close(client)

			err := client.CreateSubscriptionGroup(context.Background(),
				admin.WithGroupName(group),
			)
			if err != nil {
				panic(err)
			}
		},
	}

	cmd.Flags().StringVarP(&group, "group", "g", "", "")
	return cmd
}
