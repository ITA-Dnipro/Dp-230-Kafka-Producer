package main

import (
	"context"
	"html/template"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"parabellum.kproducer/internal/config"
	"parabellum.kproducer/internal/network/router"
	"parabellum.kproducer/internal/network/router/handler"
)

func main() {
	exitCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	// TODO: don't panic
	tmpl := template.Must(template.ParseFiles(filepath.Join(os.Getenv("PATH_TO_TEMPLATES"), "report.html")))

	app := config.NewApp()
	defer app.Close()

	mainHandler := handler.NewMainHandler(exitCtx, app.Producer, app.Grpc)
	reportHandler := handler.NewReportHandler(app.Grpc, tmpl)
	myRouter := router.NewRouter(mainHandler, reportHandler)

	go app.Start(myRouter)

	<-exitCtx.Done()
}
