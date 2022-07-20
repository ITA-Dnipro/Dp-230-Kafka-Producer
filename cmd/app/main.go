package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"parabellum.kproducer/internal/model"
	"parabellum.kproducer/internal/network"
	"parabellum.kproducer/internal/pubsub"
)
import "github.com/joho/godotenv"

var TopicName string

type AppConfig struct {
	Producer *pubsub.Producer
	Http     *network.ServerHTTP
	Grpc     *network.ClientGRPC
}

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Panicln("Error loading .env file: ", err)
	}
	TopicName = os.Getenv("KAFKA_TOPIC_API")
}

func main() {
	exitCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	app := new(AppConfig)
	app.Producer = pubsub.NewProducer(pubsub.RealKafkaWriter(os.Getenv("KAFKA_URL"), TopicName), TopicName)
	defer app.Producer.Close()
	app.Http = network.NewServerHTTP(fmt.Sprintf(":%s", os.Getenv("HTTP_ADDR")))
	app.Http.Start()
	defer app.Http.Close()
	app.Grpc = network.NewClientGRPC(exitCtx, os.Getenv("GRPC_ADDR"))
	defer app.Grpc.Close()

	var gotFromUser model.TaskFromAPI
	for {
		select {
		case gotFromUser = <-app.Http.UserQuery:
		case <-exitCtx.Done():
			log.Println("Exiting on termination signal")

			return
		}

		if len(gotFromUser.URL) == 0 ||
			len(gotFromUser.Email) == 0 ||
			len(gotFromUser.ForwardTo) == 0 {
			log.Println("User email and/or host url weren't passed")

			continue
		}

		gotFromDB, err := app.Grpc.CreateNewTask(gotFromUser)
		if err != nil {
			log.Println("Error when creating new task in DB:\t", err)
			continue
		}

		message := model.NewMessageProduce(&gotFromDB)
		err = app.Producer.PublicMessage(exitCtx, message)
		if err != nil {
			log.Printf("Error producing message [%s] to <%s>:\t%v", gotFromDB, TopicName, err)
		}
	}
}
