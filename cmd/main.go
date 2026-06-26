package main

import (
	"os"

	"github.com/echooymxq/rmq/pkg/cli/broker"
	"github.com/echooymxq/rmq/pkg/cli/cluster"
	contextcmd "github.com/echooymxq/rmq/pkg/cli/context"
	"github.com/echooymxq/rmq/pkg/cli/group"
	"github.com/echooymxq/rmq/pkg/cli/message"
	"github.com/echooymxq/rmq/pkg/cli/namesrv"
	"github.com/echooymxq/rmq/pkg/cli/topic"
	"github.com/echooymxq/rmq/pkg/config"
	"github.com/echooymxq/rmq/pkg/rocketmq"
	"github.com/spf13/cobra"
)

func main() {
	rocketmq.DisableClientLogging()

	root := &cobra.Command{
		Use:          "rmq",
		Short:        "Apache RocketMQ cli",
		SilenceUsage: true,
	}

	r := new(config.RocketMQConfig)
	r.InstallRootFlags(root)
	root.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if config.ShouldSkipResolve(cmd) {
			return nil
		}
		return r.Resolve(cmd)
	}
	root.AddCommand(contextcmd.NewCommand(r))
	root.AddCommand(cluster.NewCommand(r))
	root.AddCommand(broker.NewCommand(r))
	root.AddCommand(namesrv.NewCommand(r))
	root.AddCommand(topic.NewCommand(r))
	root.AddCommand(group.NewCommand(r))
	root.AddCommand(message.NewCommand(r))

	err := root.Execute()
	if err != nil {
		os.Exit(1)
	}

}
