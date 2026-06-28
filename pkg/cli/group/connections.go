package group

import (
	"fmt"

	"github.com/apache/rocketmq-client-go/v2/admin"
	"github.com/echooymxq/rmq/pkg/cli"
	"github.com/echooymxq/rmq/pkg/config"
	"github.com/echooymxq/rmq/pkg/rocketmq"
	"github.com/spf13/cobra"
)

func Connections(r *config.RocketMQConfig) *cobra.Command {
	var group string

	cmd := &cobra.Command{
		Use:     "connections",
		Aliases: []string{"clients", "instances"},
		Short:   "Show active consumer client connections",
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
			if len(connection.ConnectionSet) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No active consumer clients.")
				return nil
			}

			results := make([]consumerStatusResult, 0, len(connection.ConnectionSet))
			for _, client := range connection.ConnectionSet {
				runningInfo, runningErr := getConsumerRunningInfo(adminClient, group, client.ClientId, false)
				results = append(results, consumerStatusResult{
					Client:      client,
					RunningInfo: runningInfo,
					RunningErr:  runningErr,
				})
			}

			renderConsumerConnections(cmd, connection, results)
			return nil
		},
	}
	cmd.Flags().StringVarP(&group, "group", "g", "", "")
	cli.MarkFlagsRequired(cmd, "group")
	return cmd
}

func renderConsumerConnections(cmd *cobra.Command, connection *admin.ConsumerConnection, results []consumerStatusResult) {
	renderSectionTitle(cmd, "Group Summary")
	summaryTable := newSectionTable(cmd)
	summaryTable.SetHeader([]string{"Online", "InstanceCount", "ConsumeType", "MessageModel", "ConsumeFromWhere"})
	instanceCount := 0
	consumeType := "-"
	messageModel := "-"
	consumeFromWhere := "-"
	if connection != nil {
		instanceCount = len(connection.ConnectionSet)
		consumeType = valueOrDash(string(connection.ConsumeType))
		messageModel = valueOrDash(connection.MessageModel)
		consumeFromWhere = valueOrDash(connection.ConsumeFromWhere)
	}
	summaryTable.Append([]string{
		fmt.Sprintf("%t", instanceCount > 0),
		fmt.Sprintf("%d", instanceCount),
		consumeType,
		messageModel,
		consumeFromWhere,
	})
	summaryTable.Render()

	fmt.Fprintln(cmd.OutOrStdout())
	renderSectionTitle(cmd, "Consumer Instances")
	table := newSectionTable(cmd)
	table.SetHeader([]string{"ClientId", "ClientAddr", "Language", "Version"})
	for _, result := range results {
		table.Append([]string{
			result.Client.ClientId,
			result.Client.ClientAddr,
			result.Client.Language,
			consumerProperty(result.RunningInfo, "PROP_CLIENT_VERSION"),
		})
	}
	table.Render()

	subscriptions := getSubscriptions(connection)
	if len(subscriptions) > 0 {
		fmt.Fprintln(cmd.OutOrStdout())
		renderSectionTitle(cmd, "Subscriptions")
		subscriptionTable := newSectionTable(cmd)
		subscriptionTable.SetHeader([]string{"Topic", "ExpressionType", "Expression"})
		for _, subscription := range subscriptions {
			subscriptionTable.Append([]string{subscription.Topic, subscription.ExpressionType, subscription.Expression})
		}
		subscriptionTable.Render()
	}
}
