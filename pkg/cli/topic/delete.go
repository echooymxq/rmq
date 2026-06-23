package topic

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"github.com/apache/rocketmq-client-go/v2/admin"
	"github.com/echooymxq/rmq/pkg/cli"
	"github.com/echooymxq/rmq/pkg/config"
	"github.com/echooymxq/rmq/pkg/rocketmq"
	"github.com/spf13/cobra"
)

func Delete(r *config.RocketMQConfig) *cobra.Command {
	var (
		topic      string
		brokerAddr string
	)
	var cmd = &cobra.Command{
		Use:  "delete",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := rocketmq.NewAdminClient(r)
			if err != nil {
				return err
			}
			defer rocketmq.Close(client)

			if brokerAddr != "" {
				if err := deleteTopicOnBroker(client, topic, brokerAddr); err != nil {
					return err
				}
				fmt.Fprintf(cmd.OutOrStdout(), "Deleted topic %q on broker %s.\n", topic, brokerAddr)
				return nil
			}

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
				if addr != "" {
					masterAddrs = append(masterAddrs, addr)
				}
			}
			if len(masterAddrs) == 0 {
				return errors.New("no master broker found")
			}
			for _, addr := range masterAddrs {
				if err := deleteTopicOnBroker(client, topic, addr); err != nil {
					return err
				}
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&topic, "topic", "t", "", "")
	cmd.Flags().StringVarP(&brokerAddr, "brokerAddr", "b", "", "")
	cli.MarkFlagsRequired(cmd, "topic")

	return cmd
}

func deleteTopicOnBroker(client admin.Admin, topic, brokerAddr string) error {
	opts := []admin.OptionDelete{
		admin.WithTopicDelete(topic),
		admin.WithBrokerAddrDelete(brokerAddr),
	}
	if err := client.DeleteTopic(context.Background(), opts...); err != nil {
		return fmt.Errorf("delete topic %q on broker %s: %w", topic, brokerAddr, err)
	}
	return nil
}
