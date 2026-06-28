package group

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/apache/rocketmq-client-go/v2/admin"
	"github.com/echooymxq/rmq/pkg/cli"
	"github.com/echooymxq/rmq/pkg/config"
	"github.com/echooymxq/rmq/pkg/rocketmq"
	"github.com/spf13/cobra"
)

type consumerStatusResult struct {
	Client      admin.Connection
	RunningInfo *admin.ConsumerRunningInfo
	RunningErr  error
}

type runningSubscription struct {
	Topic         string
	ClassFilter   bool
	SubExpression string
}

type runningQueue struct {
	Topic                string
	Broker               string
	QueueId              int
	CommitOffset         int64
	CachedMsgMinOffset   int64
	CachedMsgMaxOffset   int64
	CachedMsgCount       int
	CachedMsgSizeInMiB   int64
	LastPullTimestamp    int64
	LastConsumeTimestamp int64
}

type runningRTTPS struct {
	Topic             string
	PullRT            float64
	PullTPS           float64
	ConsumeRT         float64
	ConsumeOKTPS      float64
	ConsumeFailedTPS  float64
	ConsumeFailedMsgs int64
}

func Status(r *config.RocketMQConfig) *cobra.Command {
	var (
		group    string
		clientId string
		stack    bool
	)

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show consumer running status",
		RunE: func(cmd *cobra.Command, args []string) error {
			adminClient, err := rocketmq.NewAdminClient(r)
			if err != nil {
				return err
			}
			defer rocketmq.Close(adminClient)

			connection, err := examineGroupConsumerConnection(adminClient, group)
			if err != nil {
				return err
			}

			clients := filterConsumerConnections(connection.ConnectionSet, clientId)
			if clientId != "" && len(clients) == 0 {
				return fmt.Errorf("consumer client %q not found in group %q", clientId, group)
			}
			if len(clients) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No active consumer clients.")
				return nil
			}

			results := make([]consumerStatusResult, 0, len(clients))
			for _, client := range clients {
				runningInfo, runningErr := getConsumerRunningInfo(adminClient, group, client.ClientId, stack)
				results = append(results, consumerStatusResult{
					Client:      client,
					RunningInfo: runningInfo,
					RunningErr:  runningErr,
				})
			}

			printConsumers(cmd, results)
			printConsumerSubscriptions(cmd, collectStatusSubscriptionRows(connection, results))
			for i, result := range results {
				printConsumerStatus(cmd, i+1, len(results), result, stack)
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&group, "group", "g", "", "")
	cmd.Flags().StringVarP(&clientId, "clientId", "c", "", "show status for one consumer client")
	cmd.Flags().BoolVarP(&stack, "stack", "s", false, "include consumer client stack dump")
	cli.MarkFlagsRequired(cmd, "group")
	return cmd
}

func filterConsumerConnections(connections []admin.Connection, clientId string) []admin.Connection {
	if clientId == "" {
		return connections
	}

	result := make([]admin.Connection, 0, 1)
	for _, connection := range connections {
		if connection.ClientId == clientId {
			result = append(result, connection)
		}
	}
	return result
}

func printConsumers(cmd *cobra.Command, results []consumerStatusResult) {
	printSectionTitle(cmd, "Consumers")
	table := newSectionTable(cmd)
	table.SetHeader([]string{"#", "ClientId", "Addr", "Version", "StartTime"})
	for i, result := range results {
		table.Append([]string{
			strconv.Itoa(i + 1),
			result.Client.ClientId,
			result.Client.ClientAddr,
			consumerProperty(result.RunningInfo, "PROP_CLIENT_VERSION"),
			consumerPropertyTime(result.RunningInfo, "PROP_CONSUMER_START_TIMESTAMP"),
		})
	}
	table.Render()
}

func printConsumerStatus(cmd *cobra.Command, index, total int, result consumerStatusResult, stack bool) {
	fmt.Fprintln(cmd.OutOrStdout())
	printConsumerHeader(cmd, index, total, result.Client)
	if result.RunningErr != nil {
		fmt.Fprintf(cmd.OutOrStdout(), "No consumer running info: %v\n", result.RunningErr)
		return
	}
	if result.RunningInfo == nil {
		fmt.Fprintln(cmd.OutOrStdout(), "No consumer running info.")
		return
	}

	queueRows := collectRunningQueueRows(result.RunningInfo)
	printConsumerQueues(cmd, queueRows)
	printConsumerRTTPS(cmd, collectRunningRTTPSRows(result.RunningInfo))
	if stack {
		printConsumerStack(cmd, result.RunningInfo.Jstack)
	}
}

func printConsumerHeader(cmd *cobra.Command, index, total int, client admin.Connection) {
	title := fmt.Sprintf("Consumer %d/%d: %s (%s)", index, total, client.ClientId, client.ClientAddr)
	fmt.Fprintln(cmd.OutOrStdout(), title)
	fmt.Fprintln(cmd.OutOrStdout(), strings.Repeat("=", len(title)))
}

func printConsumerSubscriptions(cmd *cobra.Command, rows []runningSubscription) {
	fmt.Fprintln(cmd.OutOrStdout())
	printSectionTitle(cmd, "Consumer Subscriptions")
	table := newSectionTable(cmd)
	table.SetHeader([]string{"Topic", "ClassFilter", "SubExpression"})
	for _, row := range rows {
		table.Append([]string{row.Topic, strconv.FormatBool(row.ClassFilter), row.SubExpression})
	}
	table.Render()
}

func printConsumerQueues(cmd *cobra.Command, rows []runningQueue) {
	fmt.Fprintln(cmd.OutOrStdout())
	printSectionTitle(cmd, "Consumer Queues")
	table := newSectionTable(cmd)
	table.SetHeader([]string{
		"Topic",
		"Broker",
		"QueueId",
		"ConsumerOffset",
		"CachedMsgCount",
		"CachedMsgMinOffset",
		"CachedMsgMaxOffset",
		"CachedMsgSizeInMiB",
		"LastPullTime",
		"LastConsumeTime",
	})
	for _, row := range rows {
		table.Append([]string{
			row.Topic,
			row.Broker,
			strconv.Itoa(row.QueueId),
			strconv.FormatInt(row.CommitOffset, 10),
			strconv.Itoa(row.CachedMsgCount),
			strconv.FormatInt(row.CachedMsgMinOffset, 10),
			strconv.FormatInt(row.CachedMsgMaxOffset, 10),
			strconv.FormatInt(row.CachedMsgSizeInMiB, 10),
			formatUnixMilli(row.LastPullTimestamp),
			formatUnixMilli(row.LastConsumeTimestamp),
		})
	}
	table.Render()
}

func printConsumerRTTPS(cmd *cobra.Command, rows []runningRTTPS) {
	fmt.Fprintln(cmd.OutOrStdout())
	printSectionTitle(cmd, "Consumer RT/TPS")
	table := newSectionTable(cmd)
	table.SetHeader([]string{
		"Topic",
		"PullRT",
		"PullTPS",
		"ConsumeRT",
		"ConsumeOKTPS",
		"ConsumeFailedTPS",
		"ConsumeFailedMsgsInHour",
	})
	for _, row := range rows {
		table.Append([]string{
			row.Topic,
			formatFloat(row.PullRT),
			formatFloat(row.PullTPS),
			formatFloat(row.ConsumeRT),
			formatFloat(row.ConsumeOKTPS),
			formatFloat(row.ConsumeFailedTPS),
			strconv.FormatInt(row.ConsumeFailedMsgs, 10),
		})
	}
	table.Render()
}

func printConsumerStack(cmd *cobra.Command, stack string) {
	fmt.Fprintln(cmd.OutOrStdout())
	printSectionTitle(cmd, "Consumer Stack")
	if strings.TrimSpace(stack) == "" {
		fmt.Fprintln(cmd.OutOrStdout(), "No consumer stack dump.")
		return
	}
	fmt.Fprintln(cmd.OutOrStdout(), strings.TrimRight(stack, "\r\n"))
}

func collectStatusSubscriptionRows(connection *admin.ConsumerConnection, results []consumerStatusResult) []runningSubscription {
	rowByKey := make(map[string]runningSubscription)
	add := func(topic string, classFilter bool, subExpression string) {
		row := runningSubscription{
			Topic:         valueOrDash(topic),
			ClassFilter:   classFilter,
			SubExpression: valueOrDash(subExpression),
		}
		key := row.Topic + "\x00" + strconv.FormatBool(row.ClassFilter) + "\x00" + row.SubExpression
		rowByKey[key] = row
	}

	if connection != nil {
		for topic, subscription := range connection.SubscriptionTable {
			rowTopic := subscription.Topic
			if rowTopic == "" {
				rowTopic = topic
			}
			add(rowTopic, subscription.ClassFilterMode, subscription.SubString)
		}
	}
	for _, result := range results {
		if result.RunningInfo == nil {
			continue
		}
		for _, subscription := range result.RunningInfo.SubscriptionData {
			add(subscription.Topic, subscription.ClassFilterMode, subscription.SubString)
		}
	}

	rows := make([]runningSubscription, 0, len(rowByKey))
	for _, row := range rowByKey {
		rows = append(rows, row)
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].Topic != rows[j].Topic {
			return rows[i].Topic < rows[j].Topic
		}
		if rows[i].ClassFilter != rows[j].ClassFilter {
			return !rows[i].ClassFilter
		}
		return rows[i].SubExpression < rows[j].SubExpression
	})
	return rows
}

func collectRunningQueueRows(runningInfo *admin.ConsumerRunningInfo) []runningQueue {
	if runningInfo == nil {
		return nil
	}

	rows := make([]runningQueue, 0, len(runningInfo.MQTable))
	for mq, queueInfo := range runningInfo.MQTable {
		rows = append(rows, runningQueue{
			Topic:                mq.Topic,
			Broker:               mq.BrokerName,
			QueueId:              mq.QueueId,
			CommitOffset:         queueInfo.CommitOffset,
			CachedMsgMinOffset:   queueInfo.CachedMsgMinOffset,
			CachedMsgMaxOffset:   queueInfo.CachedMsgMaxOffset,
			CachedMsgCount:       queueInfo.CachedMsgCount,
			CachedMsgSizeInMiB:   queueInfo.CachedMsgSizeInMiB,
			LastPullTimestamp:    queueInfo.LastPullTimestamp,
			LastConsumeTimestamp: queueInfo.LastConsumeTimestamp,
		})
	}

	sort.Slice(rows, func(i, j int) bool {
		if rows[i].Topic != rows[j].Topic {
			return rows[i].Topic < rows[j].Topic
		}
		if rows[i].Broker != rows[j].Broker {
			return rows[i].Broker < rows[j].Broker
		}
		return rows[i].QueueId < rows[j].QueueId
	})
	return rows
}

func collectRunningRTTPSRows(runningInfo *admin.ConsumerRunningInfo) []runningRTTPS {
	if runningInfo == nil {
		return nil
	}

	rows := make([]runningRTTPS, 0, len(runningInfo.StatusTable))
	for topic, status := range runningInfo.StatusTable {
		rows = append(rows, runningRTTPS{
			Topic:             topic,
			PullRT:            status.PullRT,
			PullTPS:           status.PullTPS,
			ConsumeRT:         status.ConsumeRT,
			ConsumeOKTPS:      status.ConsumeOKTPS,
			ConsumeFailedTPS:  status.ConsumeFailedTPS,
			ConsumeFailedMsgs: status.ConsumeFailedMsgs,
		})
	}

	sort.Slice(rows, func(i, j int) bool {
		return rows[i].Topic < rows[j].Topic
	})
	return rows
}

func formatFloat(value float64) string {
	return fmt.Sprintf("%.2f", value)
}

func consumerProperty(runningInfo *admin.ConsumerRunningInfo, key string) string {
	if runningInfo == nil {
		return "-"
	}
	return valueOrDash(runningInfo.Properties[key])
}

func consumerPropertyTime(runningInfo *admin.ConsumerRunningInfo, key string) string {
	if runningInfo == nil {
		return "-"
	}
	value := runningInfo.Properties[key]
	if value == "" {
		return "-"
	}
	timestamp, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return value
	}
	return formatUnixMilli(timestamp)
}

func formatUnixMilli(value int64) string {
	if value <= 0 {
		return "-"
	}
	return time.UnixMilli(value).Local().Format("2006-01-02 15:04:05.000")
}

func getConsumerRunningInfo(adminTool admin.Admin, group, clientId string, stack bool) (runningInfo *admin.ConsumerRunningInfo, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("get consumer running info for %q: %v", clientId, recovered)
		}
	}()
	return adminTool.GetConsumerRunningInfo(group, clientId, stack)
}
