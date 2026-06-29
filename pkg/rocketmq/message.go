package rocketmq

import "github.com/apache/rocketmq-client-go/v2/primitive"

type Message struct {
	Topic          string            `json:"topic"`
	BrokerName     string            `json:"brokerName,omitempty"`
	QueueID        *int              `json:"queueId"`
	QueueOffset    int64             `json:"queueOffset"`
	MsgID          string            `json:"msgId"`
	OffsetMsgID    string            `json:"offsetMsgId"`
	Properties     map[string]string `json:"properties,omitempty"`
	BornHost       string            `json:"bornHost,omitempty"`
	StoreHost      string            `json:"storeHost,omitempty"`
	BornTimestamp  int64             `json:"bornTimestamp,omitempty"`
	StoreTimestamp int64             `json:"storeTimestamp,omitempty"`
	ReconsumeTimes int32             `json:"reconsumeTimes,omitempty"`
}

func NewMessage(msg *primitive.MessageExt, detailed bool) Message {
	result := Message{
		Topic:       msg.Topic,
		QueueOffset: msg.QueueOffset,
		MsgID:       msg.MsgId,
		OffsetMsgID: msg.OffsetMsgId,
	}
	if msg.Queue != nil {
		result.QueueID = &msg.Queue.QueueId
		result.BrokerName = msg.Queue.BrokerName
	}
	if detailed {
		result.Properties = msg.GetProperties()
		result.BornHost = msg.BornHost
		result.StoreHost = msg.StoreHost
		result.BornTimestamp = msg.BornTimestamp
		result.StoreTimestamp = msg.StoreTimestamp
		result.ReconsumeTimes = msg.ReconsumeTimes
	}
	return result
}
