package config

import (
	"context"
	"fmt"
	"log"
	"os"

	"parabellum.kproducer/internal/network/communicator"
	"parabellum.kproducer/internal/network/server"
	"parabellum.kproducer/internal/pubsub"

	"github.com/joho/godotenv"
)

const (
	TopicSQLI = "SQLI-check"
	TopicXSS  = "XSS-check"
	Topic5XX  = "5XX-check"
)

var TopicName string

type AppConfig struct {
	Producer *pubsub.Producer
	Http     *server.HTTP
	Grpc     *communicator.ClientGRPC
}

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Panicln("Error loading .env file: ", err)
	}
	TopicName = os.Getenv("KAFKA_TOPIC_API")
}

func NewApp(ctx context.Context) *AppConfig {
	app := new(AppConfig)
	app.Grpc = communicator.NewClientGRPC(ctx, os.Getenv("GRPC_ADDR"))
	app.Producer = pubsub.NewProducer(ctx, pubsub.RealKafkaWriter(os.Getenv("KAFKA_URL"), TopicName), TopicName)
	app.Http = server.NewServerHTTP(fmt.Sprintf(":%s", os.Getenv("HTTP_ADDR")))

	return app
}

func (app *AppConfig) Start() {
	app.Http.Start()
}

func (app *AppConfig) Close() error {
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
