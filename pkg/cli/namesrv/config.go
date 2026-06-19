package namesrv

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
		namesrvAddr string
		key         string
		val         string
	)
	cmd := &cobra.Command{
		Use: "config",
		RunE: func(cmd *cobra.Command, args []string) error {
			admin, err := rocketmq.NewAdminClient(r)
			if err != nil {
				return err
			}
			defer rocketmq.Close(admin)
			// if key is not empty, update namesrv config
			if len(key) > 0 {
				err = admin.UpdateNamesrvConfig(namesrvAddr, key, val)
				if err != nil {
					return err
				}
				fmt.Println("Update namesrv config success.")
				return nil
			} else {
				// if key is empty, get namesrv config
				configs, err := admin.GetNamesrvConfig(namesrvAddr)
				if err != nil {
					return err
				}
				keys := make([]string, 0, len(configs))
				for k := range configs {
					keys = append(keys, k)
				}
				sort.Strings(keys)
				for _, k := range keys {
					fmt.Printf("%-50s=  %s\n", k, configs[k])
				}
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&namesrvAddr, "namesrvAddr", "a", "", "")
	cmd.Flags().StringVarP(&key, "key", "k", "", "")
	cmd.Flags().StringVarP(&val, "value", "v", "", "")
	cli.MarkFlagsRequired(cmd, "namesrvAddr")
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
