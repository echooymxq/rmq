package message

import (
	"github.com/echooymxq/rmq/pkg/config"
	"github.com/spf13/cobra"
)

func NewCommand(r *config.RocketMQConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "message",
		Short: "",
	}
	r.InstallRocketMQFlags(cmd)
	cmd.AddCommand(
		Query(r),
	)
	return cmd
}
