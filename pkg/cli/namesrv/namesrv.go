package namesrv

import (
	"github.com/echooymxq/rmq/pkg/config"
	"github.com/spf13/cobra"
)

func NewCommand(r *config.RocketMQConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "namesrv",
		Short: "",
	}
	r.InstallRocketMQFlags(cmd)
	cmd.AddCommand(
		Config(r),
	)
	return cmd
}
