package group

import (
	"context"
	"fmt"
	"github.com/apache/rocketmq-client-go/v2/admin"
	"github.com/echooymxq/rmq/pkg/config"
	"github.com/echooymxq/rmq/pkg/rocketmq"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"os"
	"strconv"
	"time"

	"github.com/echooymxq/rmq/pkg/cli"
)

func Describe(r *config.RocketMQConfig) *cobra.Command {
	var (
		group          string
		showConnection bool
	)

	cmd := &cobra.Command{
		Use: "describe",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := rocketmq.NewAdminClient(r)
			if err != nil {
				return err
			}
			defer rocketmq.Close(client)
			// 查看消费组客户端连接
			if showConnection {
				consumerConnectionInfo, err := client.ExamineConsumerConnectionInfo(group)
				table := tablewriter.NewWriter(os.Stdout)
				table.SetHeaderAlignment(tablewriter.ALIGN_CENTER)
				if err != nil {
					return err
				}
				table.SetHeader([]string{"ClientId", "ClientAddr", "Language", "Version"})
				table.SetAutoFormatHeaders(false)
				for _, connection := range consumerConnectionInfo.ConnectionSet {
					table.Append([]string{connection.ClientId, connection.ClientAddr, connection.Language, strconv.FormatInt(connection.Version, 10)})
				}
				table.Render()
				return nil
			}

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
	cmd.Flags().BoolVarP(&showConnection, "showConnection", "c", false, "")
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
