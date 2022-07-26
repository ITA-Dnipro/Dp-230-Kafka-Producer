package pubsub

import (
	"context"
	"encoding/json"
	"log"

	"parabellum.kproducer/internal/model"

	"github.com/segmentio/kafka-go"
)

//KafkaWriter interface mostly for test implementing purposes
type KafkaWriter interface {
	WriteMessages(context.Context, ...kafka.Message) error
	Close() error
}

//Producer structure representing message producer
type Producer struct {
	Topic       string      //topic name
	kafkaWriter KafkaWriter //writer itself
}

//RealKafkaWriter returns filled kafka.Writer from kafka-go lib
func RealKafkaWriter(url, topic string) *kafka.Writer {
	return &kafka.Writer{
		Addr:     kafka.TCP(url),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}
}

//NewProducer is a constructor for [pubsub.Producer]
func NewProducer(kwr KafkaWriter, topic string) *Producer {
	return &Producer{
		Topic:       topic,
		kafkaWriter: kwr,
	}
}

//PublicMessage sends given message to a pubsub instance of KafkaWriter into a [producer.Topic] topic
func (prod *Producer) PublicMessage(ctx context.Context, message *model.MessageProduce) error {
	valueJson, err := json.Marshal(message.Value)
	if err != nil {
		log.Printf("Error marshalling %v to json: %v\n", message.Value, err)

		return err
	}

	msg := kafka.Message{
		Key:   []byte(message.Key),
		Value: valueJson,
		Time:  message.Time,
	}

	log.Println("Publishing into Kafka topic:", prod.Topic)
	msgOut := string(msg.Value)
	if len(msgOut) > 250 {
		msgOut = msgOut[:250] + "\t..."
	}
	log.Println("\t", msgOut)

	return prod.kafkaWriter.WriteMessages(ctx, msg)
}

//Close closes producers' KafkaWriter
func (prod *Producer) Close() error {
	log.Println("Closing message producer")

	return prod.kafkaWriter.Close()
}
