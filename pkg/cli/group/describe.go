package group

import (
	"context"
	"fmt"
	"github.com/apache/rocketmq-client-go/v2/admin"
	"github.com/echooymxq/rmq/pkg/config"
	"github.com/echooymxq/rmq/pkg/rocketmq"
	"github.com/spf13/cobra"
	"time"
)

func Describe(r *config.RocketMQConfig) *cobra.Command {
	var group string

	cmd := &cobra.Command{
		Use: "describe",
		Run: func(cmd *cobra.Command, args []string) {
			adminTool, err := rocketmq.NewAdminClient(r)
			if err == nil {
				groupConfig, _ := GetSubscriptionGroupConfig(adminTool, group)
				json := groupConfig.ToJson(groupConfig, true)
				fmt.Println(json)
			}
		},
	}
	cmd.Flags().StringVarP(&group, "group", "g", "", "")
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
			}
		}
	}
	return nil, err
}
