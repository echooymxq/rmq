package topic

import (
	"context"
	"github.com/apache/rocketmq-client-go/v2/admin"
	"github.com/echooymxq/rmq/pkg/config"
	"github.com/echooymxq/rmq/pkg/rocketmq"
	"github.com/spf13/cobra"
)

func Create(r *config.RocketMQConfig) *cobra.Command {
	var (
		messageType string
		topic       string
		brokerAddr  string
	)
	var cmd = &cobra.Command{
		Use: "create",
		Run: func(cmd *cobra.Command, args []string) {
			err := r.Load()
			if err != nil {
				panic(err)
			}
			client, _ := rocketmq.NewAdminClient(r)
			defer rocketmq.Close(client)

			var opts []admin.OptionCreate
			opts = append(opts, admin.WithTopicCreate(topic))
			opts = append(opts, admin.WithAttribute("message.type", messageType))
			if brokerAddr != "" {
				opts = append(opts, admin.WithBrokerAddrCreate(brokerAddr))
			}
			_ = client.CreateTopic(context.Background(), opts...)
		},
	}

	cmd.Flags().StringVarP(&brokerAddr, "brokerAddr", "b", "", "")
	cmd.Flags().StringVarP(&topic, "topic", "t", "", "")
	cmd.Flags().StringVarP(&messageType, "messageType", "m", "NORMAL", "")

	return cmd
}
