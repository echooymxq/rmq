package contextcmd

import (
	"github.com/echooymxq/rmq/pkg/config"
	"github.com/spf13/cobra"
)

func NewCommand(r *config.RocketMQConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "context",
		Short: "Manage configuration contexts",
		Annotations: map[string]string{
			config.AnnotationSkipResolve: "true",
		},
	}
	cmd.AddCommand(
		List(r),
		Current(r),
		Use(r),
		Add(r),
	)
	return cmd
}
