package broker

import (
	"errors"
	"fmt"

	"github.com/echooymxq/rmq/pkg/cli"
	"github.com/echooymxq/rmq/pkg/config"
	"github.com/echooymxq/rmq/pkg/rocketmq"
	"github.com/spf13/cobra"
	"sort"
)

func Config(r *config.RocketMQConfig) *cobra.Command {
	var (
		brokerAddr string
		key        string
		val        string
	)
	cmd := &cobra.Command{
		Use: "config",
		RunE: func(cmd *cobra.Command, args []string) error {
			admin, err := rocketmq.NewAdminClient(r)
			if err != nil {
				return err
			}
			defer rocketmq.Close(admin)
			// if key is not empty, update broker config
			if len(key) > 0 {
				err = admin.UpdateBrokerConfig(brokerAddr, key, val)
				if err != nil {
					return err
				}
				fmt.Println("Update broker config success.")
				return nil
			} else {
				// if key is empty, get broker config
				brokerConfig, err := admin.GetBrokerConfig(brokerAddr)
				if err != nil {
					return err
				}
				keys := make([]string, 0, len(brokerConfig))
				for k := range brokerConfig {
					keys = append(keys, k)
				}
				sort.Strings(keys)
				for _, k := range keys {
					fmt.Printf("%-50s=  %s\n", k, brokerConfig[k])
				}
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&brokerAddr, "brokerAddr", "b", "", "")
	cmd.Flags().StringVarP(&key, "key", "k", "", "")
	cmd.Flags().StringVarP(&val, "value", "v", "", "")
	cli.MarkFlagsRequired(cmd, "brokerAddr")
	cli.ChainPreRunE(cmd, func(cmd *cobra.Command, args []string) error {
		if key != "" && val == "" {
			return errors.New("value is required when key is set")
		}
		if key == "" && val != "" {
			return errors.New("key is required when value is set")
		}
		return nil
	})
	return cmd
}
