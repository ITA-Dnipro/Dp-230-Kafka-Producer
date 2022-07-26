package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"parabellum.kproducer/internal/network/router/handler"
)

func NewRouter(mainHandler *handler.MainHandler, reportHandler *handler.ReportHandler) http.Handler {
	router := chi.NewRouter()
	router.Handle("/", mainHandler)
	router.Method("GET", "/{taskID}", reportHandler)

	return router
}
