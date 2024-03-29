package handler

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
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
		mh.forwardTheTask(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	http.ServeFile(w, r, os.Getenv("PATH_TO_TEMPLATES")+"index.html")
}

func (mh MainHandler) forwardTheTask(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	err := r.ParseForm()
	if err != nil {
		errMsg := fmt.Sprintf("Error getting user query\t%v", err)
		log.Println(errMsg)
		http.Error(w, errMsg, http.StatusInternalServerError)

		return
	}

	gotFromUser := getTaskFromRequest(r.PostForm)

	gotFromDB, err := mh.GrpcClient.CreateNewTask(mh.ctx, gotFromUser)
	if err != nil {
		errMsg := fmt.Sprintf("Error when creating new task in DB:\t%v", err)
		log.Println(errMsg)
		http.Error(w, errMsg, http.StatusInternalServerError)

		return
	}

	message := model.NewMessageProduce(&gotFromDB)
	err = mh.Producer.PublicMessage(mh.ctx, message)
	if err != nil {
		errMsg := fmt.Sprintf("Error producing message [%s] to <%s>:\t%v", gotFromDB, config.TopicName, err)
		log.Println(errMsg)
		http.Error(w, errMsg, http.StatusInternalServerError)

		return
	}

	linkToResult := "/" + gotFromDB.ID
	response := fmt.Sprintf("<h3><a href='%s'>%s</a></h3><p><h2><a href='/'>HOME</a></h2>", linkToResult, linkToResult)
	w.Write([]byte(response))
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
		case "SkipCrawler":
			if chk, err := strconv.ParseBool(curVal); err == nil && chk {
				result.SkipCrawler = true
			}
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
