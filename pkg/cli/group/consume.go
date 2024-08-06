package group

import (
	"context"
	"fmt"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/echooymxq/rmq/pkg/config"
	"github.com/echooymxq/rmq/pkg/rocketmq"
	"github.com/spf13/cobra"
	"os"
)

func Consume(r *config.RocketMQConfig) *cobra.Command {
	var (
		group string
		topic string
	)

	var cmd = &cobra.Command{
		Use: "consume",
		Run: func(cmd *cobra.Command, args []string) {
			sig := make(chan os.Signal)
			c, err := rocketmq.NewPushConsumer(r, group)
			defer c.Shutdown()
			err = c.Subscribe(topic, consumer.MessageSelector{},
				func(ctx context.Context, msgs ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
					for i := range msgs {
						fmt.Printf("Consume message success:%v \n", msgs[i])
					}
					return consumer.ConsumeSuccess, nil
				})
			if err != nil {
				fmt.Println(err.Error())
			}
			err = c.Start()
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(-1)
			}
			<-sig
		},
	}

	cmd.Flags().StringVarP(&group, "group", "g", "", "")
	cmd.Flags().StringVarP(&topic, "topic", "t", "", "")
	return cmd
}
