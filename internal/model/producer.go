package model

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

//TaskFromAPI task format to be received from API service
type TaskFromAPI struct {
	URL         string   `json:"url"`         //main task url to work with
	Email       string   `json:"email"`       //user email to send response back
	ForwardTo   []string `json:"forwardTo"`   //list of test-services topics names to send results to
	SkipCrawler bool     `json:"skipCrawler"` //if no crawling needed, just forward to tests
}

//TaskProduce task format to send
type TaskProduce struct {
	TaskFromAPI
	ID string `json:"id"` //main task id from DB service
}

//MessageProduce received messages representation
type MessageProduce struct {
	Key   string       //message key from pubsub provider
	Value *TaskProduce //message value
	Time  time.Time    //time of the message
}

//NewMessageProduce is a constructor for [model.MessageProduce]
func NewMessageProduce(task *TaskProduce) *MessageProduce {
	return &MessageProduce{
		Key:   fmt.Sprint(uuid.New()),
		Value: task,
		Time:  time.Now(),
	}
}
