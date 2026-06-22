package topic

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/echooymxq/rmq/pkg/cli"
	"github.com/echooymxq/rmq/pkg/config"
	"github.com/echooymxq/rmq/pkg/rocketmq"
	"github.com/spf13/cobra"
)

func Produce(r *config.RocketMQConfig) *cobra.Command {
	var (
		topic      string
		body       string
		count      int
		delayMills int64
		delayLevel int
	)
	var cmd = &cobra.Command{
		Use:   "produce",
		Short: "Produce messages to a RocketMQ topic",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if count <= 0 {
				return fmt.Errorf("count must be positive")
			}
			p, err := rocketmq.NewProducer(r, "rmq")
			if err != nil {
				return err
			}
			err = p.Start()
			if err != nil {
				return err
			}
			defer func() {
				if err := p.Shutdown(); err != nil {
					fmt.Fprintf(cmd.ErrOrStderr(), "shutdown producer: %s\n", err)
				}
			}()

			delayTimestamp := time.Now().UnixMilli() + delayMills
			for i := 0; i < count; i++ {
				msg := primitive.NewMessage(topic, []byte(body))

				if delayMills > 0 {
					msg.WithProperty("__STARTDELIVERTIME", strconv.FormatInt(delayTimestamp, 10))
				}

				// DelayLevel 1s, 5s, 10s, 30s, 1m, 2m, 3m, 4m, 5m, 6m, 7m, 8m, 9m, 10m, 20m, 30m, 1h, 2h
				// DelayLevel 的优先级高于 delayMills
				if delayLevel > 0 {
					msg.WithDelayTimeLevel(delayLevel)
				}

				result, err := p.SendSync(context.Background(), msg)
				if err != nil {
					return fmt.Errorf("produce message %d/%d: %w", i+1, count, err)
				}
				if result.Status != primitive.SendOK {
					return fmt.Errorf("produce message %d/%d failed: %s", i+1, count, result.String())
				}
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Produce %d message(s) success.\n", count)
			return nil
		},
	}

	cmd.Flags().StringVarP(&topic, "topic", "t", "", "")
	cmd.Flags().StringVarP(&body, "body", "b", "", "message body")
	cmd.Flags().IntVarP(&count, "count", "c", 1, "message count")
	cmd.Flags().Int64VarP(&delayMills, "delayMills", "m", -1, "message delay mill seconds")
	cmd.Flags().IntVarP(&delayLevel, "delayLevel", "l", -1, "message delay level")
	cli.MarkFlagsRequired(cmd, "topic")
	return cmd
}
