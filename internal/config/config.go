package config

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"parabellum.kproducer/internal/network/communicator"
	"parabellum.kproducer/internal/network/server"
	"parabellum.kproducer/internal/pubsub"
)

const (
	TopicSQLI = "SQLI-check"
	TopicXSS  = "XSS-check"
	Topic5XX  = "5XX-check"
)

var envDefaults = map[string]string{
	"KAFKA_URL":       "kafka:9092",
	"KAFKA_TOPIC_API": "API-Service-Message",
	"HTTP_ADDR":       "8888",
	"GRPC_ADDR":       "grpcserver:9090",
}

var TopicName string

type AppDependency struct {
	Producer *pubsub.Producer
	Http     *server.HTTP
	Grpc     *communicator.ClientGRPC
}

func init() {
	err := setEnvDefaults()
	if err != nil {
		log.Panicln("Error setting default env parameters")
	}
	TopicName = os.Getenv("KAFKA_TOPIC_API")
}

func setEnvDefaults() error {
	var err error
	for env, val := range envDefaults {
		if _, ok := os.LookupEnv(env); !ok {
			err = os.Setenv(env, val)
		}
		if err != nil {
			break
		}
	}

	return err
}

func NewApp() *AppDependency {
	kafkaPub := pubsub.RealKafkaWriter(os.Getenv("KAFKA_URL"), TopicName)
	result := &AppDependency{
		Producer: pubsub.NewProducer(kafkaPub, TopicName),
		Http:     server.NewServerHTTP(fmt.Sprintf(":%s", os.Getenv("HTTP_ADDR"))),
		Grpc:     communicator.NewClientGRPC(os.Getenv("GRPC_ADDR")),
	}

	return result
}

func (app *AppDependency) Start(router http.Handler) {
	app.Http.Start(router)
}

func (app *AppDependency) Close() error {
	var err error

	if app.Grpc.Close() != nil {
		err = fmt.Errorf("error closing grpc %w", err)
	}
	if app.Producer.Close() != nil {
		err = fmt.Errorf("error closing producer:\t%w", err)
	}
	if app.Http.Close() != nil {
		err = fmt.Errorf("error closing http server:\t%w", err)
	}

	return err
}
