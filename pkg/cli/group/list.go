package group

import (
	"context"

	"github.com/echooymxq/rmq/pkg/config"
	"github.com/echooymxq/rmq/pkg/rocketmq"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"os"
	"time"
)

func List(r *config.RocketMQConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use: "list",
		RunE: func(cmd *cobra.Command, args []string) error {
			admin, err := rocketmq.NewAdminClient(r)
			if err != nil {
				return err
			}
			defer rocketmq.Close(admin)

			groupNames := make(map[string]struct{})
			clusterInfo, err := admin.ExamineBrokerClusterInfo()
			if err != nil {
				return err
			}
			for _, brokerData := range clusterInfo.BrokerAddrTable {
				brokerAddr := brokerData.BrokerAddresses[rocketmq.MasterId]
				subscriptionGroupWrapper, err := admin.GetAllSubscriptionGroup(context.Background(), brokerAddr, 3*time.Second)
				if err != nil {
					return err
				}
				for groupName := range subscriptionGroupWrapper.SubscriptionGroupTable {
					groupNames[groupName] = struct{}{}
				}
			}
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"Name"})
			for groupName := range groupNames {
				table.Append([]string{groupName})
			}
			table.Render()
			return nil
		},
	}
	return cmd
}
