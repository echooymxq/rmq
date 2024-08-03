package rocketmq

import (
	"fmt"
	"github.com/apache/rocketmq-client-go/v2/admin"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/echooymxq/rmq/pkg/config"
)

func NewAdminClient(r *config.RocketMQConfig) (admin.Admin, error) {
	namesrv, accessKey, secretKey := r.GetNamesrvAddrs(), r.AccessKey, r.SecretKey
	return admin.NewAdmin(
		admin.WithResolver(primitive.NewPassthroughResolver(namesrv)),
		admin.WithCredentials(primitive.Credentials{
			AccessKey: accessKey,
			SecretKey: secretKey,
		}),
	)
}

func Close(admin admin.Admin) {
	err := admin.Close()
	if err != nil {
		fmt.Printf("Shutdown admin error: %s", err.Error())
	}
}
