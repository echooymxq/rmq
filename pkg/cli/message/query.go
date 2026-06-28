package message

import (
	"fmt"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/apache/rocketmq-client-go/v2/admin"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/echooymxq/rmq/pkg/cli"
	"github.com/echooymxq/rmq/pkg/config"
	"github.com/echooymxq/rmq/pkg/rocketmq"
	"github.com/spf13/cobra"
)

const (
	messageStatusConsumed    = "consumed"
	messageStatusNotConsumed = "not_consumed"
	messageStatusUnknown     = "unknown"
)

func Query(r *config.RocketMQConfig) *cobra.Command {
	var (
		topic     string
		messageId string
		group     string
	)
	cmd := &cobra.Command{
		Use:   "query -t TOPIC -m MESSAGE_ID [-g GROUP]",
		Short: "Query a message by MessageId",
		RunE: func(cmd *cobra.Command, args []string) error {
			adminClient, err := rocketmq.NewAdminClient(r)
			if err != nil {
				return err
			}
			defer rocketmq.Close(adminClient)

			messageExt, err := queryMessage(adminClient, topic, messageId)
			if err != nil {
				return err
			}
			consumeStatus := ""
			if group != "" {
				stats, err := adminClient.ExamineTopicConsumeStats(group, messageExt.Topic)
				if err != nil {
					return err
				}
				consumeStatus = calculateMessageStatus(messageExt, stats)
			}
			printMessage(cmd, messageExt, consumeStatus)
			return nil
		},
	}
	cmd.Flags().StringVarP(&topic, "topic", "t", "", "message topic")
	cmd.Flags().StringVarP(&messageId, "messageId", "m", "", "MessageId")
	cmd.Flags().StringVarP(&group, "group", "g", "", "consumer group")
	cli.MarkFlagsRequired(cmd, "topic", "messageId")
	return cmd
}

func queryMessage(adminClient admin.Admin, topic, messageID string) (*primitive.MessageExt, error) {
	messageExt, viewErr := adminClient.ViewMessage(messageID)
	if viewErr == nil {
		if messageExt == nil {
			return nil, fmt.Errorf("query message by MessageId returned empty message, MessageId: %s", messageID)
		}
		if messageExt.Topic == topic {
			return messageExt, nil
		}

		uniqueMessageExt, uniqueErr := adminClient.QueryMessageByUniqKey(topic, messageID)
		if uniqueErr == nil {
			return uniqueMessageExt, nil
		}
		return nil, fmt.Errorf("message %q found in topic %q, want topic %q", messageID, messageExt.Topic, topic)
	}

	messageExt, uniqueErr := adminClient.QueryMessageByUniqKey(topic, messageID)
	if uniqueErr == nil {
		return messageExt, nil
	}
	return nil, fmt.Errorf("query message by MessageId failed, topic: %s, MessageId: %s", topic, messageID)
}

func printMessage(cmd *cobra.Command, msg *primitive.MessageExt, consumeStatus string) {
	out := cmd.OutOrStdout()
	if consumeStatus == "" {
		fmt.Fprintln(out, "Message")
	} else {
		fmt.Fprintf(out, "Message (ConsumeStatus: %s)\n", consumeStatus)
	}
	fmt.Fprintln(out, "-------")
	fmt.Fprintf(out, "Topic: %s\n", msg.Topic)
	fmt.Fprintf(out, "MessageId: %s\n", msg.MsgId)
	fmt.Fprintf(out, "Broker: %s\n", msg.Queue.BrokerName)
	fmt.Fprintf(out, "QueueId: %d\n", msg.Queue.QueueId)
	fmt.Fprintf(out, "QueueOffset: %d\n", msg.QueueOffset)
	fmt.Fprintf(out, "BornTimestamp: %d\n", msg.BornTimestamp)
	fmt.Fprintf(out, "StoreTimestamp: %d\n", msg.StoreTimestamp)
	fmt.Fprintf(out, "ReconsumeTimes: %d\n", msg.ReconsumeTimes)
	fmt.Fprintf(out, "BodySize: %d\n", len(msg.Body))
	printMessageBody(out, msg.Body)
	printMessageProperties(out, msg.GetProperties())
}

func printMessageBody(out interface {
	Write([]byte) (int, error)
}, body []byte) {
	bodyText := formatMessageBody(body)
	if strings.ContainsAny(bodyText, "\r\n") {
		fmt.Fprintln(out, "Body:")
		fmt.Fprintln(out, bodyText)
		return
	}
	fmt.Fprintf(out, "Body: %s\n", bodyText)
}

func formatMessageBody(body []byte) string {
	if len(body) == 0 {
		return "<empty>"
	}
	if !utf8.Valid(body) {
		return fmt.Sprintf("<binary %d bytes>", len(body))
	}
	text := string(body)
	if !isPrintableBody(text) {
		return fmt.Sprintf("<binary %d bytes>", len(body))
	}

	const maxBodyRunes = 1024
	runes := []rune(text)
	if len(runes) <= maxBodyRunes {
		return text
	}
	return fmt.Sprintf("%s... <truncated, %d bytes total>", string(runes[:maxBodyRunes]), len(body))
}

func isPrintableBody(text string) bool {
	for _, r := range text {
		if r == '\n' || r == '\r' || r == '\t' {
			continue
		}
		if !unicode.IsPrint(r) {
			return false
		}
	}
	return true
}

func printMessageProperties(out interface {
	Write([]byte) (int, error)
}, properties map[string]string) {
	if len(properties) == 0 {
		return
	}
	fmt.Fprintln(out, "Properties:")
	keys := make([]string, 0, len(properties))
	for key := range properties {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		fmt.Fprintf(out, "  %s: %s\n", key, properties[key])
	}
}

func calculateMessageStatus(msg *primitive.MessageExt, stats *admin.ConsumeStats) string {
	if msg.Queue == nil || stats == nil {
		return messageStatusUnknown
	}

	offset, ok := stats.OffsetTable[primitive.MessageQueue{
		Topic:      msg.Topic,
		BrokerName: msg.Queue.BrokerName,
		QueueId:    msg.Queue.QueueId,
	}]
	if !ok {
		return messageStatusUnknown
	}

	if offset.ConsumerOffset > msg.QueueOffset {
		return messageStatusConsumed
	}

	return messageStatusNotConsumed
}
