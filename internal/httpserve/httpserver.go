package httpserve

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	kproducer "parabellum.kproducer"
	"parabellum.kproducer/internal/model"
)

type ServerHTTP struct {
	Server          *http.Server
	UserQuery       chan model.TaskFromAPI
	shutdownTimeout time.Duration
	middlewares     []http.Handler
}

func (srv *ServerHTTP) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, m := range srv.middlewares {
		m.ServeHTTP(w, r)
	}

	srv.serveMainPage(w, r)
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
				result.ForwardTo = append(result.ForwardTo, kproducer.TopicSQLI)
			}
		case "TestXSS":
			if chk, err := strconv.ParseBool(curVal); err == nil && chk {
				result.ForwardTo = append(result.ForwardTo, kproducer.TopicXSS)
			}
		case "Test5XX":
			if chk, err := strconv.ParseBool(curVal); err == nil && chk {
				result.ForwardTo = append(result.ForwardTo, kproducer.Topic5XX)
			}
		}
	}

	return result
}

func (srv *ServerHTTP) serveMainPage(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			log.Println("error getting user query\t", err)
			_, _ = fmt.Fprintln(w, "Some error occurred")

			return
		}

		srv.UserQuery <- getTaskFromRequest(r.PostForm)
	}

	w.Header().Set("Content-Type", "text/html")
	http.ServeFile(w, r, "./templates/index.html")
}

func New(addr string, midl ...http.Handler) *ServerHTTP {
	cert, err := tls.LoadX509KeyPair("server.crt", "server.key")
	var tlsConf *tls.Config
	if err != nil {
		log.Println("Error getting tls-certificate")
	} else {
		tlsConf = new(tls.Config)
		tlsConf.Certificates = []tls.Certificate{cert}
	}

	return &ServerHTTP{
		Server: &http.Server{
			Addr:      addr,
			TLSConfig: tlsConf,
		},
		middlewares:     midl,
		UserQuery:       make(chan model.TaskFromAPI, 10),
		shutdownTimeout: 5 * time.Second,
	}
}

func (srv *ServerHTTP) SetShutdownTimeout(t time.Duration) {
	srv.shutdownTimeout = t
}

func (srv *ServerHTTP) Start() {
	go func() {
		var err error
		srv.Server.Handler = srv
		if srv.Server.TLSConfig == nil {
			err = srv.Server.ListenAndServe()
		} else {
			err = srv.Server.ListenAndServeTLS("server.crt", "server.key")
		}
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("server listen&serve error: %v\n", err)
		}
	}()
	log.Println("Starting http-server on addr:\t", srv.Server.Addr)
}

func (srv *ServerHTTP) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), srv.shutdownTimeout)
	defer cancel()
	log.Println("Shutting down http-server")
	close(srv.UserQuery)

	return srv.Server.Shutdown(ctx)
}
