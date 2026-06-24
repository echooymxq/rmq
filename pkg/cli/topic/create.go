package topic

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
		messageType string
		topic       string
		brokerAddr  string
		queueNum    int
	)
	var cmd = &cobra.Command{
		Use:  "create",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			queueNums, err := resolveCreateQueueNum(cmd, queueNum)
			if err != nil {
				return err
			}

			client, err := rocketmq.NewAdminClient(r)
			if err != nil {
				return err
			}
			defer rocketmq.Close(client)

			if brokerAddr != "" {
				if err := createTopicOnBroker(client, topic, messageType, brokerAddr, queueNums); err != nil {
					return err
				}
				fmt.Fprintf(cmd.OutOrStdout(), "Created topic %q on broker %s.\n", topic, brokerAddr)
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
				if err := createTopicOnBroker(client, topic, messageType, addr, queueNums); err != nil {
					return err
				}
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Created topic %q on %d broker(s): %s.\n", topic, len(masterAddrs), strings.Join(masterAddrs, ","))
			return nil
		},
	}

	cmd.Flags().StringVarP(&brokerAddr, "brokerAddr", "b", "", "")
	cmd.Flags().StringVarP(&topic, "topic", "t", "", "")
	cmd.Flags().StringVarP(&messageType, "messageType", "m", "NORMAL", "")
	cmd.Flags().IntVarP(&queueNum, "queueNum", "q", 0, "read and write queue count")
	cli.MarkFlagsRequired(cmd, "topic")

	return cmd
}

func resolveCreateQueueNum(cmd *cobra.Command, queueNum int) (int, error) {
	if !cmd.Flags().Changed("queueNum") {
		return 0, nil
	}
	if queueNum <= 0 {
		return 0, errors.New("queueNum must be positive")
	}

	return queueNum, nil
}

func createTopicOnBroker(client admin.Admin, topic, messageType, brokerAddr string, queueNum int) error {
	opts := []admin.OptionCreate{
		admin.WithTopicCreate(topic),
		admin.WithAttribute("message.type", messageType),
		admin.WithBrokerAddrCreate(brokerAddr),
	}
	if queueNum > 0 {
		opts = append(opts,
			admin.WithReadQueueNums(queueNum),
			admin.WithWriteQueueNums(queueNum),
		)
	}
	if err := client.CreateTopic(context.Background(), opts...); err != nil {
		return fmt.Errorf("create topic %q on broker %s: %w", topic, brokerAddr, err)
	}
	return nil
}
