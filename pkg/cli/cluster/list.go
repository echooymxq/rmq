package cluster

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/echooymxq/rmq/pkg/config"
	"github.com/echooymxq/rmq/pkg/rocketmq"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

func List(r *config.RocketMQConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List RocketMQ cluster brokers",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			adminClient, err := rocketmq.NewAdminClient(r)
			if err != nil {
				return err
			}
			defer rocketmq.Close(adminClient)

			clusterInfo, err := adminClient.ExamineBrokerClusterInfo()
			if err != nil {
				return fmt.Errorf("examine broker cluster info: %w", err)
			}

			table := tablewriter.NewWriter(cmd.OutOrStdout())
			table.SetHeader([]string{"Cluster", "Broker Group", "Broker ID", "Role", "Address"})

			// Sort cluster names, broker groups, and broker IDs for stable output.
			clusterNames := make([]string, 0, len(clusterInfo.ClusterAddrTable))
			for clusterName := range clusterInfo.ClusterAddrTable {
				clusterNames = append(clusterNames, clusterName)
			}
			sort.Strings(clusterNames)

			for _, clusterName := range clusterNames {
				brokerNames := append([]string(nil), clusterInfo.ClusterAddrTable[clusterName]...)
				sort.Strings(brokerNames)

				for _, brokerName := range brokerNames {
					addresses := clusterInfo.BrokerAddrTable[brokerName].BrokerAddresses
					brokerIDs := make([]int64, 0, len(addresses))
					for brokerID := range addresses {
						brokerIDs = append(brokerIDs, brokerID)
					}
					sort.Slice(brokerIDs, func(i, j int) bool {
						return brokerIDs[i] < brokerIDs[j]
					})

					for _, brokerID := range brokerIDs {
						table.Append([]string{
							clusterName,
							brokerName,
							strconv.FormatInt(brokerID, 10),
							brokerRole(brokerID),
							addresses[brokerID],
						})
					}
				}
			}
			table.Render()
			return nil
		},
	}
	return cmd
}

func brokerRole(brokerID int64) string {
	if brokerID == rocketmq.MasterId {
		return "MASTER"
	}
	// RocketMQ treats non-zero broker IDs as slave brokers.
	return "SLAVE"
}
