package namesrv

import (
	"fmt"
	"github.com/echooymxq/rmq/pkg/config"
	"github.com/echooymxq/rmq/pkg/rocketmq"
	"github.com/spf13/cobra"
	"sort"
)

func Config(r *config.RocketMQConfig) *cobra.Command {
	var (
		namesrvAddr string
		key         string
		val         string
	)
	cmd := &cobra.Command{
		Use: "config",
		Run: func(cmd *cobra.Command, args []string) {
			admin, err := rocketmq.NewAdminClient(r)
			defer rocketmq.Close(admin)
			// if key is not empty, update namesrv config
			if len(key) > 0 {
				err = admin.UpdateNamesrvConfig(namesrvAddr, key, val)
				if err != nil {
					fmt.Println(err)
					return
				}
				fmt.Println("Update namesrv config success.")
			} else {
				// if key is empty, get namesrv config
				configs, err := admin.GetNamesrvConfig(namesrvAddr)
				if err == nil {
					keys := make([]string, 0, len(configs))
					for k := range configs {
						keys = append(keys, k)
					}
					sort.Strings(keys)
					for _, k := range keys {
						fmt.Printf("%-50s=  %s\n", k, configs[k])
					}
				}
			}
		},
	}
	cmd.Flags().StringVarP(&namesrvAddr, "namesrvAddr", "n", "", "")
	cmd.Flags().StringVarP(&key, "key", "k", "", "")
	cmd.Flags().StringVarP(&val, "value", "v", "", "")
	return cmd
}
