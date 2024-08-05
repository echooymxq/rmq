package topic

import (
	"context"
	"fmt"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/echooymxq/rmq/pkg/config"
	"github.com/echooymxq/rmq/pkg/rocketmq"
	"github.com/spf13/cobra"
	"strconv"
	"time"
)

func Produce(r *config.RocketMQConfig) *cobra.Command {
	var (
		topic      string
		delayMills int64
		delayLevel int
	)
	var cmd = &cobra.Command{
		Use: "produce",
		Run: func(cmd *cobra.Command, args []string) {
			p, err := rocketmq.NewProducer(r, "rmq")
			err = p.Start()
			defer p.Shutdown()
			if err == nil {
				msg := primitive.Message{
					Topic: topic,
				}

				if delayMills > 0 {
					msg.WithProperty("__STARTDELIVERTIME", strconv.FormatInt(time.Now().UnixMilli()+delayMills, 10))
				}

				// DelayLevel 1s, 5s, 10s, 30s, 1m, 2m, 3m, 4m, 5m, 6m, 7m, 8m, 9m, 10m, 20m, 30m, 1h, 2h
				// DelayLevel 的优先级高于 delayMills
				if delayLevel > 0 {
					msg.WithDelayTimeLevel(delayLevel)
				}

				result, err := p.SendSync(context.Background(), &msg)
				if err != nil {
					fmt.Println(err)
					return
				}
				if result.Status == primitive.SendOK {
					fmt.Printf("Produce message success, :%s\n", result.String())
				}
			}
		},
	}

	cmd.Flags().StringVarP(&topic, "topic", "t", "", "")
	cmd.Flags().Int64VarP(&delayMills, "delayMills", "m", -1, "message delay mill seconds")
	cmd.Flags().IntVarP(&delayLevel, "delayLevel", "l", -1, "message delay level")
	return cmd
}
