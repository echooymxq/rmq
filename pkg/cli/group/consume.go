package group

import (
	"context"
	"encoding/json"
	"os"
	"sync"

	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/echooymxq/rmq/pkg/cli"
	"github.com/echooymxq/rmq/pkg/config"
	"github.com/echooymxq/rmq/pkg/rocketmq"
	"github.com/spf13/cobra"
)

func Consume(r *config.RocketMQConfig) *cobra.Command {
	var (
		group   string
		topic   string
		verbose bool
	)

	var cmd = &cobra.Command{
		Use: "consume",
		RunE: func(cmd *cobra.Command, args []string) error {
			sig := make(chan os.Signal)
			c, err := rocketmq.NewPushConsumer(r, group)
			if err != nil {
				return err
			}
			defer c.Shutdown()

			encoder := json.NewEncoder(cmd.OutOrStdout())
			var encoderMu sync.Mutex
			err = c.Subscribe(topic, consumer.MessageSelector{},
				func(ctx context.Context, msgs ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
					for i := range msgs {
						encoderMu.Lock()
						if err := encoder.Encode(rocketmq.NewMessage(msgs[i], verbose)); err != nil {
							encoderMu.Unlock()
							return consumer.ConsumeRetryLater, err
						}
						encoderMu.Unlock()
					}
					return consumer.ConsumeSuccess, nil
				})
			if err != nil {
				return err
			}
			err = c.Start()
			if err != nil {
				return err
			}
			<-sig
			return nil
		},
	}

	cmd.Flags().StringVarP(&group, "group", "g", "", "")
	cmd.Flags().StringVarP(&topic, "topic", "t", "", "")
	cmd.Flags().BoolVar(&verbose, "verbose", false, "")
	cli.MarkFlagsRequired(cmd, "group", "topic")
	return cmd
}
