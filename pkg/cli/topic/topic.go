package topic

import (
	"github.com/echooymxq/rmq/pkg/config"
	"github.com/spf13/cobra"
)

func NewCommand(r *config.RocketMQConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "topic",
		Short: "Create, delete, produce to and consume RocketMQ topics",
	}
	r.InstallRocketMQFlags(cmd)

	cmd.AddCommand(
		Create(r),
	)
	return cmd
}
