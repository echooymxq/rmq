package group

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/apache/rocketmq-client-go/v2/admin"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/echooymxq/rmq/pkg/cli"
	"github.com/echooymxq/rmq/pkg/config"
	"github.com/echooymxq/rmq/pkg/rocketmq"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

func Lag(r *config.RocketMQConfig) *cobra.Command {
	var (
		group string
		topic string
	)

	cmd := &cobra.Command{
		Use:   "lag",
		Short: "Show consumer lag by queue",
		RunE: func(cmd *cobra.Command, args []string) error {
			adminClient, err := rocketmq.NewAdminClient(r)
			if err != nil {
				return err
			}
			defer rocketmq.Close(adminClient)

			stats, err := examineGroupConsumeStats(adminClient, group, topic)
			if err != nil {
				return err
			}

			lags := calculateConsumeLags(stats)
			renderConsumeLag(cmd, lags)
			return nil
		},
	}
	cmd.Flags().StringVarP(&group, "group", "g", "", "")
	cmd.Flags().StringVarP(&topic, "topic", "t", "", "")
	cli.MarkFlagsRequired(cmd, "group")
	return cmd
}

func examineGroupConsumeStats(adminTool admin.Admin, group, topic string) (*admin.ConsumeStats, error) {
	if topic != "" {
		return safeExamineTopicConsumeStats(adminTool, group, topic)
	}

	// Prefer the broker-side aggregate query when topic is omitted. Some broker
	// or SDK versions cannot handle an empty topic, so fall back to per-topic stats.
	stats, err := safeExamineTopicConsumeStats(adminTool, group, "")
	if err == nil {
		return stats, nil
	}

	connection, connErr := examineGroupConsumerConnection(adminTool, group)
	if connErr != nil {
		return nil, err
	}

	subscriptions := getSubscriptions(connection)
	if len(subscriptions) == 0 {
		return nil, err
	}

	// Use the consumer group's subscriptions to discover topics, query each
	// topic separately, then merge the offsets into one group-level result.
	consumeStats := &admin.ConsumeStats{
		OffsetTable: make(map[primitive.MessageQueue]admin.OffsetWrapper),
	}
	var lastErr error
	for _, subscription := range subscriptions {
		topicStats, topicErr := safeExamineTopicConsumeStats(adminTool, group, subscription.Topic)
		if topicErr != nil {
			lastErr = topicErr
			continue
		}
		mergeConsumeStats(consumeStats, topicStats)
	}

	if len(consumeStats.OffsetTable) == 0 && lastErr != nil {
		return nil, lastErr
	}
	return consumeStats, nil
}

func safeExamineTopicConsumeStats(adminTool admin.Admin, group, topic string) (stats *admin.ConsumeStats, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("examine topic consume stats: %v", recovered)
		}
	}()
	return adminTool.ExamineTopicConsumeStats(group, topic)
}

func mergeConsumeStats(dst, src *admin.ConsumeStats) {
	if dst == nil || src == nil {
		return
	}
	if dst.OffsetTable == nil {
		dst.OffsetTable = make(map[primitive.MessageQueue]admin.OffsetWrapper)
	}
	for mq, offset := range src.OffsetTable {
		dst.OffsetTable[mq] = offset
	}
	dst.ConsumeTps += src.ConsumeTps
	dst.ConsumeByteRate += src.ConsumeByteRate
}

type consumeLag struct {
	Topic          string
	Broker         string
	QueueID        int
	ConsumerOffset int64
	BrokerOffset   int64
	Lag            int64
	LastTimestamp  int64
}

func calculateConsumeLags(stats *admin.ConsumeStats) []consumeLag {
	if stats == nil {
		return nil
	}

	lags := make([]consumeLag, 0, len(stats.OffsetTable))
	for mq, offset := range stats.OffsetTable {
		lag := offset.BrokerOffset - offset.ConsumerOffset
		if lag < 0 {
			lag = 0
		}

		lags = append(lags, consumeLag{
			Topic:          mq.Topic,
			Broker:         mq.BrokerName,
			QueueID:        mq.QueueId,
			ConsumerOffset: offset.ConsumerOffset,
			BrokerOffset:   offset.BrokerOffset,
			Lag:            lag,
			LastTimestamp:  offset.LastTimestamp,
		})
	}

	sort.Slice(lags, func(i, j int) bool {
		if lags[i].Lag != lags[j].Lag {
			return lags[i].Lag > lags[j].Lag
		}
		if lags[i].Topic != lags[j].Topic {
			return lags[i].Topic < lags[j].Topic
		}
		if lags[i].Broker != lags[j].Broker {
			return lags[i].Broker < lags[j].Broker
		}
		return lags[i].QueueID < lags[j].QueueID
	})
	return lags
}

func renderConsumeLag(cmd *cobra.Command, lags []consumeLag) {
	table := tablewriter.NewWriter(cmd.OutOrStdout())
	table.SetHeader([]string{"Topic", "Broker", "QueueId", "ConsumerOffset", "BrokerOffset", "Lag", "LastTimestamp"})
	table.SetAutoFormatHeaders(false)
	for _, lag := range lags {
		table.Append([]string{
			lag.Topic,
			lag.Broker,
			strconv.Itoa(lag.QueueID),
			strconv.FormatInt(lag.ConsumerOffset, 10),
			strconv.FormatInt(lag.BrokerOffset, 10),
			strconv.FormatInt(lag.Lag, 10),
			formatUnixMilli(lag.LastTimestamp),
		})
	}
	table.Render()
}
