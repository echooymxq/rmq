package group

import (
	"context"
	"github.com/echooymxq/rmq/pkg/config"
	"github.com/echooymxq/rmq/pkg/rocketmq"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"log"
	"os"
	"time"
)

func List(r *config.RocketMQConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use: "list",
		Run: func(cmd *cobra.Command, args []string) {
			admin, err := rocketmq.NewAdminClient(r)
			defer rocketmq.Close(admin)
			if err == nil {
				groupNames := make(map[string]struct{})
				clusterInfo, err := admin.ExamineBrokerClusterInfo()
				if err == nil {
					for _, brokerData := range clusterInfo.BrokerAddrTable {
						brokerAddr := brokerData.BrokerAddresses[rocketmq.MasterId]
						subscriptionGroupWrapper, err := admin.GetAllSubscriptionGroup(context.Background(), brokerAddr, 3*time.Second)
						if err == nil {
							for groupName, _ := range subscriptionGroupWrapper.SubscriptionGroupTable {
								groupNames[groupName] = struct{}{}
							}
						} else {
							log.Fatalln(err)
						}
					}
					table := tablewriter.NewWriter(os.Stdout)
					table.SetHeader([]string{"Name"})
					for groupName := range groupNames {
						table.Append([]string{groupName})
					}
					table.Render()
				}
			}
		},
	}
	return cmd
}
