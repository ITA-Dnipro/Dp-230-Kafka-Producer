package handlers

import (
	"context"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"parabellum.kproducer/internal/config"
	"parabellum.kproducer/internal/model"
	"parabellum.kproducer/internal/network/communicator"
	"parabellum.kproducer/internal/pubsub"
)

type MainHandler struct {
	Producer   *pubsub.Producer
	GrpcClient *communicator.ClientGRPC

	ctx context.Context
}

func NewMainHandler(ctx context.Context, producer *pubsub.Producer, client *communicator.ClientGRPC) *MainHandler {
	return &MainHandler{
		Producer:   producer,
		GrpcClient: client,
		ctx:        ctx,
	}
}

func (mh MainHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		mh.forwardTheTask(r)
	}

	w.Header().Set("Content-Type", "text/html")
	http.ServeFile(w, r, "./templates/index.html")
}

func (mh MainHandler) forwardTheTask(r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Println("Error getting user query\t", err)

		return
	}

	gotFromUser := getTaskFromRequest(r.PostForm)

	gotFromDB, err := mh.GrpcClient.CreateNewTask(mh.ctx, gotFromUser)
	if err != nil {
		log.Println("Error when creating new task in DB:\t", err)

		return
	}

	message := model.NewMessageProduce(&gotFromDB)
	err = mh.Producer.PublicMessage(mh.ctx, message)
	if err != nil {
		log.Printf("Error producing message [%s] to <%s>:\t%v", gotFromDB, config.TopicName, err)

		return
	}
}

func getTaskFromRequest(reqValues url.Values) model.TaskFromAPI {
	result := model.TaskFromAPI{}
	for k := range reqValues {
		curVal := reqValues.Get(k)
		switch k {
		case "LoginEmail":
			result.Email = curVal
		case "HostName":
			result.URL = curVal
		case "TestSQLI":
			if chk, err := strconv.ParseBool(curVal); err == nil && chk {
				result.ForwardTo = append(result.ForwardTo, config.TopicSQLI)
			}
		case "TestXSS":
			if chk, err := strconv.ParseBool(curVal); err == nil && chk {
				result.ForwardTo = append(result.ForwardTo, config.TopicXSS)
			}
		case "Test5XX":
			if chk, err := strconv.ParseBool(curVal); err == nil && chk {
				result.ForwardTo = append(result.ForwardTo, config.Topic5XX)
			}
		}
	}

	return result
}
