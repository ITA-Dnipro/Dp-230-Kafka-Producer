package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"parabellum.kproducer/internal/model"
	"parabellum.kproducer/internal/pubsub"
)
import "github.com/joho/godotenv"

var TopicName string

type Config struct {
	Producer *pubsub.Producer
}

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Panicln("Error loading .env file: ", err)
	}
	TopicName = os.Getenv("KAFKA_TOPIC_API")
}

func main() {
	exitCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	app := new(Config)
	app.Producer = pubsub.NewProducer(pubsub.RealKafkaWriter(os.Getenv("KAFKA_URL"), TopicName), TopicName)
	defer app.Producer.Close()

	for {
		//TODO: get task from API over RPC
		stubFromAPI := model.TaskFromAPI{
			URL:       "http://httpstat.us/",
			ForwardTo: []string{"SQLI-check"},
		}

		//TODO: send data to Mongo over DBservice & receive task ID
		stubFromDB := model.TaskProduce{
			TaskFromAPI: stubFromAPI,
			ID:          "main-task-db-id-1",
		}

		//TODO: compose task & send it to corresponding topic
		message := model.NewMessageProduce(&stubFromDB)
		err := app.Producer.PublicMessage(exitCtx, message)
		if err != nil {
			log.Printf("Error producing message [%s] to <%s>:\t%v", stubFromDB, TopicName, err)
		}

		select {
		case <-exitCtx.Done():
			log.Println("Exiting on termination signal")

			return
		default:
		}
	}
}
