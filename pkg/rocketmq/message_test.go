package rocketmq

import (
	"encoding/json"
	"testing"

	"github.com/apache/rocketmq-client-go/v2/primitive"
)

func TestNewMessageCoreFields(t *testing.T) {
	queueID := 1
	msg := &primitive.MessageExt{
		Message: primitive.Message{
			Topic: "TopicTest",
			Queue: &primitive.MessageQueue{QueueId: queueID},
		},
		MsgId:       "msg-id",
		OffsetMsgId: "offset-msg-id",
		QueueOffset: 42,
	}

	got := NewMessage(msg, false)
	if got.Topic != msg.Topic {
		t.Fatalf("Topic = %q, want %q", got.Topic, msg.Topic)
	}
	if got.MsgID != msg.MsgId {
		t.Fatalf("MsgID = %q, want %q", got.MsgID, msg.MsgId)
	}
	if got.OffsetMsgID != msg.OffsetMsgId {
		t.Fatalf("OffsetMsgID = %q, want %q", got.OffsetMsgID, msg.OffsetMsgId)
	}
	if got.QueueID == nil || *got.QueueID != queueID {
		t.Fatalf("QueueID = %v, want %d", got.QueueID, queueID)
	}
	if got.QueueOffset != msg.QueueOffset {
		t.Fatalf("QueueOffset = %d, want %d", got.QueueOffset, msg.QueueOffset)
	}
}

func TestNewMessageDetailedFields(t *testing.T) {
	msg := &primitive.MessageExt{
		Message: primitive.Message{
			Topic: "TopicTest",
			Queue: &primitive.MessageQueue{QueueId: 2},
		},
		MsgId:          "msg-id",
		OffsetMsgId:    "offset-msg-id",
		QueueOffset:    43,
		BornHost:       "127.0.0.1:61268",
		StoreHost:      "127.0.0.1:10912",
		BornTimestamp:  1781846343265,
		StoreTimestamp: 1781846343291,
		ReconsumeTimes: 3,
	}
	msg.WithProperties(map[string]string{"KEYS": "order-1"})

	got := NewMessage(msg, true)
	if got.Properties["KEYS"] != "order-1" {
		t.Fatalf("Properties[KEYS] = %q, want order-1", got.Properties["KEYS"])
	}
	if got.BornHost != msg.BornHost || got.StoreHost != msg.StoreHost {
		t.Fatalf("hosts = %q/%q, want %q/%q", got.BornHost, got.StoreHost, msg.BornHost, msg.StoreHost)
	}
	if got.BornTimestamp != msg.BornTimestamp || got.StoreTimestamp != msg.StoreTimestamp {
		t.Fatalf("timestamps = %d/%d, want %d/%d", got.BornTimestamp, got.StoreTimestamp, msg.BornTimestamp, msg.StoreTimestamp)
	}
	if got.ReconsumeTimes != msg.ReconsumeTimes {
		t.Fatalf("ReconsumeTimes = %d, want %d", got.ReconsumeTimes, msg.ReconsumeTimes)
	}
}

func TestNewMessageNilQueue(t *testing.T) {
	msg := &primitive.MessageExt{
		Message:     primitive.Message{Topic: "TopicTest"},
		MsgId:       "msg-id",
		OffsetMsgId: "offset-msg-id",
	}

	got := NewMessage(msg, false)
	if got.QueueID != nil {
		t.Fatalf("QueueID = %v, want nil", got.QueueID)
	}
}

func TestMessageJSONCoreFields(t *testing.T) {
	queueID := 1
	msg := &primitive.MessageExt{
		Message: primitive.Message{
			Topic: "TopicTest",
			Queue: &primitive.MessageQueue{QueueId: queueID},
		},
		MsgId:       "msg-id",
		OffsetMsgId: "offset-msg-id",
		QueueOffset: 42,
	}

	bytes, err := json.Marshal(NewMessage(msg, false))
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(bytes, &got); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	want := map[string]any{
		"topic":       "TopicTest",
		"msgId":       "msg-id",
		"offsetMsgId": "offset-msg-id",
		"queueId":     float64(queueID),
		"queueOffset": float64(42),
	}
	if len(got) != len(want) {
		t.Fatalf("fields = %v, want only %v", got, want)
	}
	for key, wantValue := range want {
		if got[key] != wantValue {
			t.Fatalf("%s = %v, want %v", key, got[key], wantValue)
		}
	}
}

func TestMessageJSONDetailedFields(t *testing.T) {
	msg := &primitive.MessageExt{
		Message: primitive.Message{
			Topic: "TopicTest",
			Queue: &primitive.MessageQueue{QueueId: 2},
		},
		MsgId:          "msg-id",
		OffsetMsgId:    "offset-msg-id",
		QueueOffset:    43,
		BornHost:       "127.0.0.1:61268",
		StoreHost:      "127.0.0.1:10912",
		BornTimestamp:  1781846343265,
		StoreTimestamp: 1781846343291,
		ReconsumeTimes: 3,
	}
	msg.WithProperties(map[string]string{"KEYS": "order-1"})

	bytes, err := json.Marshal(NewMessage(msg, true))
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(bytes, &got); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if got["bornHost"] != msg.BornHost || got["storeHost"] != msg.StoreHost {
		t.Fatalf("hosts = %v/%v, want %q/%q", got["bornHost"], got["storeHost"], msg.BornHost, msg.StoreHost)
	}
	properties, ok := got["properties"].(map[string]any)
	if !ok {
		t.Fatalf("properties = %T, want object", got["properties"])
	}
	if properties["KEYS"] != "order-1" {
		t.Fatalf("properties.KEYS = %v, want order-1", properties["KEYS"])
	}
}
