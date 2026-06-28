package group

import (
	"context"
	"fmt"
	"time"

	"github.com/apache/rocketmq-client-go/v2/admin"
	"github.com/spf13/cobra"

	"github.com/echooymxq/rmq/pkg/cli"
	"github.com/echooymxq/rmq/pkg/config"
	"github.com/echooymxq/rmq/pkg/rocketmq"
)

func Describe(r *config.RocketMQConfig) *cobra.Command {
	var group string

	cmd := &cobra.Command{
		Use: "describe",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := rocketmq.NewAdminClient(r)
			if err != nil {
				return err
			}
			defer rocketmq.Close(client)

			// 查看消费组配置
			groupConfig, err := GetSubscriptionGroupConfig(client, group)
			if err != nil {
				return err
			}
			if groupConfig == nil {
				return fmt.Errorf("subscription group %q not found", group)
			}
			json := groupConfig.ToJson(groupConfig, true)
			fmt.Println(json)
			return nil
		},
	}
	cmd.Flags().StringVarP(&group, "group", "g", "", "")
	cli.MarkFlagsRequired(cmd, "group")
	return cmd
}

func GetSubscriptionGroupConfig(adminTool admin.Admin, group string) (*admin.SubscriptionGroupConfig, error) {
	clusterInfo, err := adminTool.ExamineBrokerClusterInfo()
	if err == nil {
		for _, brokerData := range clusterInfo.BrokerAddrTable {
			brokerAddr := brokerData.BrokerAddresses[rocketmq.MasterId]
			subscriptionGroupWrapper, err := adminTool.GetAllSubscriptionGroup(context.Background(), brokerAddr, 3*time.Second)
			if err == nil {
				for groupName, groupConfig := range subscriptionGroupWrapper.SubscriptionGroupTable {
					if groupName == group {
						return &groupConfig, nil
					}
				}
			} else {
				return nil, err
			}
		}
	}
	return nil, err
}
