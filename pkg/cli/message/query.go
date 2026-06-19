package message

import (
	"fmt"

	"github.com/echooymxq/rmq/pkg/cli"
	"github.com/echooymxq/rmq/pkg/config"
	"github.com/echooymxq/rmq/pkg/rocketmq"
	"github.com/spf13/cobra"
)

func Query(r *config.RocketMQConfig) *cobra.Command {
	var (
		topic     string
		messageId string
	)
	cmd := &cobra.Command{
		Use: "query",
		RunE: func(cmd *cobra.Command, args []string) error {
			admin, err := rocketmq.NewAdminClient(r)
			if err != nil {
				return err
			}
			defer rocketmq.Close(admin)
			messageExt, err := admin.ViewMessage(messageId)
			if err != nil {
				return err
			}
			fmt.Printf("query message success: %s\n", messageExt.String())
			return nil
		},
	}
	cmd.Flags().StringVarP(&topic, "topic", "t", "", "")
	cmd.Flags().StringVarP(&messageId, "msgId", "m", "", "offsetMsgId")
	cli.MarkFlagsRequired(cmd, "msgId")
	return cmd
}
