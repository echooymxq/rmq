package group

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/apache/rocketmq-client-go/v2/admin"
	"github.com/echooymxq/rmq/pkg/cli"
	"github.com/echooymxq/rmq/pkg/config"
	"github.com/echooymxq/rmq/pkg/rocketmq"
	"github.com/spf13/cobra"
)

func Delete(r *config.RocketMQConfig) *cobra.Command {
	var group string

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a subscription group and its consumer offsets",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := rocketmq.NewAdminClient(r)
			if err != nil {
				return err
			}
			defer rocketmq.Close(client)

			masterAddrs, err := getMasterBrokerAddrs(client)
			if err != nil {
				return err
			}

			for _, addr := range masterAddrs {
				if err := client.DeleteSubscriptionGroup(context.Background(), addr, group, true); err != nil {
					return fmt.Errorf("delete subscription group %q on broker %s: %w", group, addr, err)
				}
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Deleted consumer group %q and offsets on %d broker(s): %s.\n", group, len(masterAddrs), strings.Join(masterAddrs, ", "))
			return nil
		},
	}

	cmd.Flags().StringVarP(&group, "group", "g", "", "")
	cli.MarkFlagsRequired(cmd, "group")
	return cmd
}

func getMasterBrokerAddrs(adminTool admin.Admin) ([]string, error) {
	clusterInfo, err := adminTool.ExamineBrokerClusterInfo()
	if err != nil {
		return nil, fmt.Errorf("examine broker cluster info: %w", err)
	}

	brokerNames := make([]string, 0, len(clusterInfo.BrokerAddrTable))
	for brokerName := range clusterInfo.BrokerAddrTable {
		brokerNames = append(brokerNames, brokerName)
	}
	sort.Strings(brokerNames)

	masterAddrs := make([]string, 0, len(brokerNames))
	for _, brokerName := range brokerNames {
		addr := clusterInfo.BrokerAddrTable[brokerName].BrokerAddresses[rocketmq.MasterId]
		if addr != "" {
			masterAddrs = append(masterAddrs, addr)
		}
	}
	if len(masterAddrs) == 0 {
		return nil, errors.New("no master broker found")
	}
	return masterAddrs, nil
}
