package topic

import (
	"context"
	"fmt"
	"github.com/echooymxq/rmq/pkg/config"
	"github.com/echooymxq/rmq/pkg/rocketmq"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"os"
)

func List(r *config.RocketMQConfig) *cobra.Command {
	var cmd = &cobra.Command{
		Use: "list",
		Run: func(cmd *cobra.Command, args []string) {
			admin, err := rocketmq.NewAdminClient(r)
			defer rocketmq.Close(admin)
			result, err := admin.FetchAllTopicList(context.Background())
			if err != nil {
				fmt.Println("FetchAllTopicList error:", err.Error())
			}
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"Name"})
			for i := range result.TopicList {
				topic := result.TopicList[i]
				table.Append([]string{topic})
			}
			table.Render()
		},
	}
	return cmd
}
