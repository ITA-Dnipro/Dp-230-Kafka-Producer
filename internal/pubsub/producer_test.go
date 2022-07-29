package pubsub

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/segmentio/kafka-go"

	"parabellum.kproducer/internal/model"
)

const topicName = "some-topic"

var errOnContext = fmt.Errorf("exit on context")
var errWrongConversion = fmt.Errorf("wrong message conversion")

var producedMessage = &model.MessageProduce{
	Key: "some-key",
	Value: &model.TaskProduce{
		ID: "some-task-id",
		TaskFromAPI: model.TaskFromAPI{
			URL:         "https://some-url.com",
			Email:       "some@mail.com",
			ForwardTo:   nil,
			SkipCrawler: false,
		},
	},
	Time: time.Now(),
}

type StubProducer struct {
	KafkaWriter
	errorToReturn error
}

func (st *StubProducer) WriteMessages(ctx context.Context, msg ...kafka.Message) error {
	select {
	case <-ctx.Done():
		return errOnContext
	default:
		if len(msg) > 0 {
			if !bytes.Equal(msg[0].Key, []byte(producedMessage.Key)) {
				return errWrongConversion
			}
			tojson, _ := json.Marshal(producedMessage.Value)
			if !bytes.Equal(msg[0].Value, tojson) {
				return errWrongConversion
			}
			if !msg[0].Time.Equal(producedMessage.Time) {
				return errWrongConversion
			}
		}
	}

	return st.errorToReturn
}

func (st *StubProducer) Close() error {
	return st.errorToReturn
}

func TestProducer_PublicMessage(t *testing.T) {
	ctxCancel, cancel := context.WithCancel(context.Background())
	cancel()

	tests := []struct {
		Name        string
		Writer      KafkaWriter
		ExpectError error
		ctx         context.Context
		message     *model.MessageProduce
	}{
		{
			Name:        "success",
			Writer:      &StubProducer{errorToReturn: nil},
			ExpectError: nil,
			ctx:         context.Background(),
			message:     producedMessage,
		},
		{
			Name:        "exit on context",
			Writer:      &StubProducer{errorToReturn: nil},
			ExpectError: errOnContext,
			ctx:         ctxCancel,
			message:     producedMessage,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			producer := NewProducer(test.Writer, topicName)
			resErr := producer.PublicMessage(test.ctx, test.message)
			producer.Close()
			if resErr != test.ExpectError {
				t.Errorf("expected:%v, actual:%v", test.ExpectError, resErr)
			}
		})
	}
}
