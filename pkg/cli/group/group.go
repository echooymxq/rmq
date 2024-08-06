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
		Create(r),
		List(r),
		Consume(r),
		Describe(r),
	)
	return cmd
}
