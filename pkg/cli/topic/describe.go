package topic

import (
	"context"
	"fmt"
	"github.com/apache/rocketmq-client-go/v2/admin"
	"github.com/echooymxq/rmq/pkg/config"
	"github.com/echooymxq/rmq/pkg/rocketmq"
	"github.com/spf13/cobra"
)

func Describe(r *config.RocketMQConfig) *cobra.Command {
	var (
		topic string
	)
	var cmd = &cobra.Command{
		Use: "describe",
		Run: func(cmd *cobra.Command, args []string) {
			adminClient, err := rocketmq.NewAdminClient(r)
			defer rocketmq.Close(adminClient)
			topicRouteData, err := adminClient.ExamineTopicRouteInfo(context.Background(), topic)
			var configs []*admin.TopicConfig
			if err == nil {
				for _, brokerData := range topicRouteData.BrokerDataList {
					addr := brokerData.BrokerAddresses[rocketmq.MasterId]
					topicConfig, err := adminClient.ExamineTopicConfig(context.Background(), addr, topic)
					if err == nil {
						configs = append(configs, topicConfig)
					} else {
						println(err)
					}
				}
			}
			if len(configs) > 0 {
				topicConfig := configs[0]
				fmt.Println(topicConfig.ToJson(topicConfig, true))
			}
		},
	}

	cmd.Flags().StringVarP(&topic, "topic", "t", "", "")
	return cmd
}
