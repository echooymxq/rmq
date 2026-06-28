package group

import (
	"fmt"
	"sort"
	"strings"

	"github.com/apache/rocketmq-client-go/v2/admin"
	"github.com/echooymxq/rmq/pkg/rocketmq"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

func examineGroupConsumerConnection(adminTool admin.Admin, group string) (*admin.ConsumerConnection, error) {
	clusterInfo, err := adminTool.ExamineBrokerClusterInfo()
	if err != nil {
		return nil, err
	}

	merged := &admin.ConsumerConnection{}
	seenConnections := make(map[string]map[string]struct{})
	var (
		foundBroker bool
		foundInfo   bool
		lastErr     error
	)

	for _, brokerData := range clusterInfo.BrokerAddrTable {
		brokerAddr := brokerData.BrokerAddresses[rocketmq.MasterId]
		if brokerAddr == "" {
			continue
		}
		foundBroker = true

		connection, err := adminTool.ExamineBrokerConsumerConnectionInfo(brokerAddr, group)
		if err != nil {
			if isConsumerOfflineError(err) {
				continue
			}
			lastErr = err
			continue
		}

		foundInfo = true
		aggregateConsumerConnection(merged, connection, seenConnections)
	}

	if !foundBroker {
		return nil, fmt.Errorf("no master broker found")
	}
	if !foundInfo && lastErr != nil {
		return nil, lastErr
	}
	sortConsumerConnections(merged.ConnectionSet)
	return merged, nil
}

// aggregateConsumerConnection folds one broker's consumer view into the group-level view.
// A group can be visible from multiple brokers, so client entries must be de-duplicated.
func aggregateConsumerConnection(dst, src *admin.ConsumerConnection, seenConnections map[string]map[string]struct{}) {
	if dst == nil || src == nil {
		return
	}

	for _, connection := range src.ConnectionSet {
		// The same consumer instance can be reported by multiple brokers.
		// ClientId and ClientAddr identify the instance; language/version are display fields.
		clientAddrs := seenConnections[connection.ClientId]
		if clientAddrs == nil {
			clientAddrs = make(map[string]struct{})
			seenConnections[connection.ClientId] = clientAddrs
		}
		if _, ok := clientAddrs[connection.ClientAddr]; ok {
			continue
		}
		clientAddrs[connection.ClientAddr] = struct{}{}
		dst.ConnectionSet = append(dst.ConnectionSet, connection)
	}

	if len(src.SubscriptionTable) > 0 {
		if dst.SubscriptionTable == nil {
			dst.SubscriptionTable = src.SubscriptionTable
		} else {
			for topic, subscription := range src.SubscriptionTable {
				dst.SubscriptionTable[topic] = subscription
			}
		}
	}

	if dst.ConsumeType == "" {
		dst.ConsumeType = src.ConsumeType
	}
	if dst.MessageModel == "" {
		dst.MessageModel = src.MessageModel
	}
	if dst.ConsumeFromWhere == "" {
		dst.ConsumeFromWhere = src.ConsumeFromWhere
	}
}

func isConsumerOfflineError(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "not online") ||
		strings.Contains(message, "consumer_not_online") ||
		strings.Contains(message, "no consumer")
}

func sortConsumerConnections(connections []admin.Connection) {
	sort.Slice(connections, func(i, j int) bool {
		if connections[i].ClientId != connections[j].ClientId {
			return connections[i].ClientId < connections[j].ClientId
		}
		return connections[i].ClientAddr < connections[j].ClientAddr
	})
}

type subscription struct {
	Topic          string
	ExpressionType string
	Expression     string
}

func getSubscriptions(connection *admin.ConsumerConnection) []subscription {
	if connection == nil {
		return nil
	}

	subscriptions := make([]subscription, 0, len(connection.SubscriptionTable))
	for topic, subscriptionData := range connection.SubscriptionTable {
		rowTopic := subscriptionData.Topic
		if rowTopic == "" {
			rowTopic = topic
		}

		subscriptions = append(subscriptions, subscription{
			Topic:          rowTopic,
			ExpressionType: valueOrDash(subscriptionData.ExpType),
			Expression:     valueOrDash(subscriptionData.SubString),
		})
	}

	sort.Slice(subscriptions, func(i, j int) bool {
		if subscriptions[i].Topic != subscriptions[j].Topic {
			return subscriptions[i].Topic < subscriptions[j].Topic
		}
		if subscriptions[i].ExpressionType != subscriptions[j].ExpressionType {
			return subscriptions[i].ExpressionType < subscriptions[j].ExpressionType
		}
		return subscriptions[i].Expression < subscriptions[j].Expression
	})
	return subscriptions
}

func printSectionTitle(cmd *cobra.Command, title string) {
	fmt.Fprintln(cmd.OutOrStdout(), title)
	fmt.Fprintln(cmd.OutOrStdout(), strings.Repeat("-", len(title)))
}

func newSectionTable(cmd *cobra.Command) *tablewriter.Table {
	table := tablewriter.NewWriter(cmd.OutOrStdout())
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAutoFormatHeaders(false)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetHeaderLine(false)
	table.SetBorder(false)
	return table
}

func valueOrDash(value string) string {
	if value == "" {
		return "-"
	}
	return value
}
