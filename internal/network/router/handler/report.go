package handler

import (
	"context"
	"fmt"
	"html/template"
	"net/http"

	"github.com/go-chi/chi/v5"

	"parabellum.kproducer/internal/model"
)

type ReportClient interface {
	GetReport(ctx context.Context, id string) (*model.Report, error)
}

type ReportHandler struct {
	reportClient ReportClient
	tmpl         *template.Template
}

func NewReportHandler(rc ReportClient, tmpl *template.Template) *ReportHandler {
	return &ReportHandler{reportClient: rc, tmpl: tmpl}
}

func (hs *ReportHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "taskID")
	report, err := hs.reportClient.GetReport(r.Context(), id)
	if err != nil {
		//TODO: map errors
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := hs.tmpl.Execute(w, report); err != nil {
		//TODO: handle error
		fmt.Println("report handle error:", err)
	}
}
