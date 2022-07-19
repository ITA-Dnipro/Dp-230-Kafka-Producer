package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"parabellum.kproducer/internal/httpserve"
	"parabellum.kproducer/internal/model"
	"parabellum.kproducer/internal/pubsub"
)
import "github.com/joho/godotenv"

var TopicName string

type Config struct {
	Producer *pubsub.Producer
	Http     *httpserve.ServerHTTP
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

	app := new(Config)

	app.Producer = pubsub.NewProducer(pubsub.RealKafkaWriter(os.Getenv("KAFKA_URL"), TopicName), TopicName)
	defer app.Producer.Close()

	app.Http = httpserve.New(fmt.Sprintf(":%s", os.Getenv("HTTP_PORT")))
	app.Http.Start()
	defer app.Http.Close()

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

		//TODO: send data to Mongo over DBservice & receive task ID
		stubFromDB := model.TaskProduce{
			TaskFromAPI: gotFromUser,
			ID:          "main-task-db-id-1",
		}

		message := model.NewMessageProduce(&stubFromDB)
		err := app.Producer.PublicMessage(exitCtx, message)
		if err != nil {
			log.Printf("Error producing message [%s] to <%s>:\t%v", stubFromDB, TopicName, err)
		}
	}
}
