package message

import (
	"fmt"
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
		Run: func(cmd *cobra.Command, args []string) {
			admin, err := rocketmq.NewAdminClient(r)
			defer rocketmq.Close(admin)
			if err == nil {
				messageExt, err := admin.ViewMessage(messageId)
				if err == nil {
					fmt.Printf("query message success: %s\n", messageExt.String())
				} else {
					fmt.Println(err)
				}
			}
		},
	}
	cmd.Flags().StringVarP(&topic, "topic", "t", "", "")
	cmd.Flags().StringVarP(&messageId, "msgId", "m", "", "offsetMsgId")
	return cmd
}
