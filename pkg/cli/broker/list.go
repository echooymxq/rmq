package broker

import (
	"fmt"
	"github.com/echooymxq/rmq/pkg/config"
	"github.com/echooymxq/rmq/pkg/rocketmq"
	"github.com/spf13/cobra"
)

func List(r *config.RocketMQConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use: "list",
		Run: func(cmd *cobra.Command, args []string) {
			admin, err := rocketmq.NewAdminClient(r)
			defer rocketmq.Close(admin)
			clusterInfo, err := admin.ExamineBrokerClusterInfo()
			if err == nil {
				fmt.Printf("%-16s  %-22s  %-4s  %-22s\n",
					"#Cluster Name",
					"#Broker Name",
					"#BID",
					"#Addr",
				)
				for clusterName, brokerNames := range clusterInfo.ClusterAddrTable {
					for _, brokerName := range brokerNames {
						for brokerId, brokerAddr := range clusterInfo.BrokerAddrTable[brokerName].BrokerAddresses {
							fmt.Printf("%-16s  %-22s  %-4d  %-22s\n",
								clusterName,
								brokerName,
								brokerId,
								brokerAddr,
							)
						}
					}
				}
			}
		},
	}
	return cmd
}
