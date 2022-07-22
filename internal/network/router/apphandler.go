package router

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"parabellum.kproducer/internal/config"
	"parabellum.kproducer/internal/model"
	"parabellum.kproducer/internal/network/communicator"
	"parabellum.kproducer/internal/network/server"
	"parabellum.kproducer/internal/pubsub"

	"github.com/go-chi/chi/v5"
)

type AppHandler struct {
	Producer   *pubsub.Producer
	GrpcClient *communicator.ClientGRPC
	HttpServer *server.HTTP
	router     chi.Router
	ctx        context.Context
}

func ConfigureHandler(ctx context.Context, app *config.AppConfig) {
	result := &AppHandler{
		Producer:   app.Producer,
		GrpcClient: app.Grpc,
		HttpServer: app.Http,
		ctx:        ctx,
	}

	result.initServerHandler()
}

func (hs *AppHandler) initServerHandler() {
	hs.router = chi.NewRouter()
	hs.router.Get("/", hs.serveMainPage)
	hs.router.Post("/", hs.serveMainPage)
	hs.router.Get("/{taskID}", hs.returnReport)

	hs.HttpServer.SetRouter(hs.router)
}

func (hs *AppHandler) returnReport(w http.ResponseWriter, r *http.Request) {
	//TODO: implement result output here
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, "Your report <b>#%s</b> is generated.\n", chi.URLParam(r, "taskID"))
	fmt.Fprintln(w, "<p><a href='/'>HOME</a>")
}

func (hs *AppHandler) serveMainPage(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		hs.forwardTheTask(r)
	}

	w.Header().Set("Content-Type", "text/html")
	http.ServeFile(w, r, "./templates/index.html")
}

func (hs *AppHandler) forwardTheTask(r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Println("Error getting user query\t", err)

		return
	}

	gotFromUser := getTaskFromRequest(r.PostForm)

	gotFromDB, err := hs.GrpcClient.CreateNewTask(gotFromUser)
	if err != nil {
		log.Println("Error when creating new task in DB:\t", err)

		return
	}

	message := model.NewMessageProduce(&gotFromDB)
	err = hs.Producer.PublicMessage(message)
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
