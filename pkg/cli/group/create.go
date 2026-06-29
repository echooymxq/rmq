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

func Create(r *config.RocketMQConfig) *cobra.Command {
	var (
		group string
	)
	var cmd = &cobra.Command{
		Use: "create",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := rocketmq.NewAdminClient(r)
			if err != nil {
				return err
			}
			defer rocketmq.Close(client)

			clusterInfo, err := client.ExamineBrokerClusterInfo()
			if err != nil {
				return fmt.Errorf("examine broker cluster info: %w", err)
			}

			brokerNames := make([]string, 0, len(clusterInfo.BrokerAddrTable))
			for brokerName := range clusterInfo.BrokerAddrTable {
				brokerNames = append(brokerNames, brokerName)
			}
			sort.Strings(brokerNames)

			var masterAddrs []string
			for _, brokerName := range brokerNames {
				addr := clusterInfo.BrokerAddrTable[brokerName].BrokerAddresses[rocketmq.MasterId]
				if addr == "" {
					continue
				}
				masterAddrs = append(masterAddrs, addr)
			}
			if len(masterAddrs) == 0 {
				return errors.New("no master broker found")
			}

			for _, addr := range masterAddrs {
				err = client.CreateSubscriptionGroup(context.Background(),
					admin.WithGroupName(group),
					admin.WithBrokerAddr(addr),
				)
				if err != nil {
					return fmt.Errorf("create subscription group %q on broker %s: %w", group, addr, err)
				}
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Created consumer group %q on %d broker(s): %s.\n", group, len(masterAddrs), strings.Join(masterAddrs, ","))
			return nil
		},
	}

	cmd.Flags().StringVarP(&group, "group", "g", "", "")
	cli.MarkFlagsRequired(cmd, "group")
	return cmd
}
