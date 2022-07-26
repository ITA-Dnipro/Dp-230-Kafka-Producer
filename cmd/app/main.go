package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"parabellum.kproducer/internal/config"
	"parabellum.kproducer/internal/network/router"
	"parabellum.kproducer/internal/network/router/handlers"
)

func main() {
	exitCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	app := config.NewApp()
	defer app.Close()
	myRouter := router.NewRouter(handlers.NewMainHandler(exitCtx, app.Producer, app.Grpc), handlers.NewReportHandler())

	go app.Start(myRouter)

	<-exitCtx.Done()
}
