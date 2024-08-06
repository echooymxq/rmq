package main

import (
	"github.com/echooymxq/rmq/pkg/cli/group"
	"github.com/echooymxq/rmq/pkg/cli/message"
	"github.com/echooymxq/rmq/pkg/cli/topic"
	"github.com/echooymxq/rmq/pkg/config"
	"github.com/spf13/cobra"
	"os"
)

func main() {
	root := &cobra.Command{
		Use:   "rmq",
		Short: "Apache RocketMQ cli",
	}

	r := new(config.RocketMQConfig)
	pf := root.PersistentFlags()
	pf.StringVar(&r.ConfigFile, "config", "", "config file")
	root.AddCommand(topic.NewCommand(r))
	root.AddCommand(group.NewCommand(r))
	root.AddCommand(message.NewCommand(r))

	err := root.Execute()
	if err != nil {
		os.Exit(1)
	}

}
