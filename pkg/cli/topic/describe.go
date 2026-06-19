package topic

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/apache/rocketmq-client-go/v2/admin"
	"github.com/olekukonko/tablewriter"
	"os"
	"strconv"

	"github.com/echooymxq/rmq/pkg/cli"
	"github.com/echooymxq/rmq/pkg/config"
	"github.com/echooymxq/rmq/pkg/rocketmq"
	"github.com/spf13/cobra"
)

func Describe(r *config.RocketMQConfig) *cobra.Command {
	var (
		topic string
		route bool
		stats bool
	)
	var cmd = &cobra.Command{
		Use: "describe",
		RunE: func(cmd *cobra.Command, args []string) error {
			adminClient, err := rocketmq.NewAdminClient(r)
			if err != nil {
				return err
			}
			defer rocketmq.Close(adminClient)
			topicRouteData, err := adminClient.ExamineTopicRouteInfo(context.Background(), topic)
			if err != nil {
				return fmt.Errorf("examine topic route info: %w", err)
			}
			// If route is true, get the topic route info
			if route {
				bytes, _ := json.MarshalIndent(topicRouteData, "", "  ")
				fmt.Println(string(bytes))
				return nil
			}
			// If stats is true, get the topic stats
			if stats {
				topicStats, err := adminClient.ExamineTopicStats(topic)
				if err != nil {
					return fmt.Errorf("examine topic stats: %w", err)
				}
				table := tablewriter.NewWriter(os.Stdout)
				table.SetAlignment(tablewriter.ALIGN_CENTER)
				table.SetBorders(tablewriter.Border{Left: true, Top: true, Right: true, Bottom: true})
				table.SetAutoFormatHeaders(false)
				table.SetHeader([]string{"BrokerName", "QueueId", "MinOffset", "MaxOffset", "LastUpdateTime"})
				for mq, topicOffset := range topicStats.OffsetTable {
					table.Append([]string{
						mq.BrokerName,
						strconv.Itoa(mq.QueueId),
						strconv.FormatInt(topicOffset.MinOffset, 10),
						strconv.FormatInt(topicOffset.MaxOffset, 10),
						strconv.FormatInt(topicOffset.LastUpdateTimestamp, 10),
					})
				}
				table.Render()
				return nil
			}
			var configs []*admin.TopicConfig
			for _, brokerData := range topicRouteData.BrokerDataList {
				addr := brokerData.BrokerAddresses[rocketmq.MasterId]
				topicConfig, err := adminClient.ExamineTopicConfig(context.Background(), addr, topic)
				if err != nil {
					return fmt.Errorf("examine topic config: %w", err)
				}
				configs = append(configs, topicConfig)
			}
			if len(configs) > 0 {
				topicConfig := configs[0]
				fmt.Println(topicConfig.ToJson(topicConfig, true))
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&topic, "topic", "t", "", "")
	cmd.Flags().BoolVarP(&route, "route", "r", false, "Show topic route info")
	cmd.Flags().BoolVarP(&stats, "stats", "s", false, "Show topic stats")
	cli.MarkFlagsRequired(cmd, "topic")
	return cmd
}

type QueueData struct {
	ReadQueueNums  int
	WriteQueueNums int
	Perm           int
}
