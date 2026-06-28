package group

import (
	"github.com/echooymxq/rmq/pkg/config"
	"github.com/spf13/cobra"
)

func NewCommand(r *config.RocketMQConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "group",
		Short: "",
	}
	r.InstallRocketMQFlags(cmd)

	cmd.AddCommand(
		Connections(r),
		Create(r),
		Consume(r),
		Delete(r),
		Describe(r),
		Lag(r),
		List(r),
		Status(r),
	)
	return cmd
}
