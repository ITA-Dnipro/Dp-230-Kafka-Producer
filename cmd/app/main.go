package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"parabellum.kproducer/internal/config"
	"parabellum.kproducer/internal/network/router"
)

func main() {
	exitCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	app := config.NewApp(exitCtx)
	router.ConfigureHandler(exitCtx, app)
	defer app.Close()

	go app.Start()

	<-exitCtx.Done()
}
