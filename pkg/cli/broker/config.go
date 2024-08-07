package broker

import (
	"fmt"
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
		Run: func(cmd *cobra.Command, args []string) {
			admin, err := rocketmq.NewAdminClient(r)
			defer rocketmq.Close(admin)
			// if key is not empty, update broker config
			if len(key) > 0 {
				err = admin.UpdateBrokerConfig(brokerAddr, key, val)
				if err != nil {
					fmt.Println(err)
					return
				}
				fmt.Println("Update broker config success.")
			} else {
				// if key is empty, get broker config
				brokerConfig, err := admin.GetBrokerConfig(brokerAddr)
				if err == nil {
					keys := make([]string, 0, len(brokerConfig))
					for k := range brokerConfig {
						keys = append(keys, k)
					}
					sort.Strings(keys)
					for _, k := range keys {
						fmt.Printf("%-50s=  %s\n", k, brokerConfig[k])
					}
				}
			}
		},
	}
	cmd.Flags().StringVarP(&brokerAddr, "brokerAddr", "b", "", "")
	cmd.Flags().StringVarP(&key, "key", "k", "", "")
	cmd.Flags().StringVarP(&val, "value", "v", "", "")
	return cmd
}
