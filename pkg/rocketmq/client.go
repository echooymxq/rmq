package rocketmq

import (
	"fmt"
	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/admin"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/producer"
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

func NewProducer(r *config.RocketMQConfig, group string) (rocketmq.Producer, error) {
	namesrv, accessKey, secretKey := r.GetNamesrvAddrs(), r.AccessKey, r.SecretKey
	return rocketmq.NewProducer(
		producer.WithNsResolver(primitive.NewPassthroughResolver(namesrv)),
		producer.WithRetry(2),
		producer.WithGroupName(group),
		producer.WithTrace(&primitive.TraceConfig{
			Access:   primitive.Local,
			Resolver: primitive.NewPassthroughResolver(namesrv),
		}),
		producer.WithCredentials(primitive.Credentials{
			AccessKey: accessKey,
			SecretKey: secretKey,
		}),
	)
}

func NewPushConsumer(r *config.RocketMQConfig, group string) (rocketmq.PushConsumer, error) {
	namesrv, accessKey, secretKey := r.GetNamesrvAddrs(), r.AccessKey, r.SecretKey
	return rocketmq.NewPushConsumer(
		consumer.WithGroupName(group),
		consumer.WithNsResolver(primitive.NewPassthroughResolver(namesrv)),
		consumer.WithCredentials(primitive.Credentials{
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
