package handler

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type ReportHandler struct {
}

func NewReportHandler() *ReportHandler {
	return &ReportHandler{}
}

func (hs ReportHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	//TODO: implement result output here
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, "Your report <b>#%s</b> is generated.\n", chi.URLParam(r, "taskID"))
	fmt.Fprintln(w, "<p><a href='/'>HOME</a>")
}
