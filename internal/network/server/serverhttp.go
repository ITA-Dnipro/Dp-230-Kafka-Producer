package server

import (
	"context"
	"crypto/tls"
	"log"
	"net/http"
	"time"
)

type HTTP struct {
	Server          *http.Server
	shutdownTimeout time.Duration
}

func NewServerHTTP(addr string) *HTTP {
	cert, err := tls.LoadX509KeyPair("server.crt", "server.key")
	var tlsConf *tls.Config
	if err != nil {
		log.Println("Error getting tls-certificate")
	} else {
		tlsConf = new(tls.Config)
		tlsConf.Certificates = []tls.Certificate{cert}
	}

	return &HTTP{
		Server: &http.Server{
			Addr:              addr,
			TLSConfig:         tlsConf,
			ReadHeaderTimeout: time.Second,
		},
		shutdownTimeout: 5 * time.Second,
	}
}

func (srv *HTTP) SetRouter(r http.Handler) {
	srv.Server.Handler = r
}

func (srv *HTTP) SetShutdownTimeout(t time.Duration) {
	srv.shutdownTimeout = t
}

func (srv *HTTP) Start() {
	var err error
	log.Println("Starting http-server on addr:\t", srv.Server.Addr)
	if srv.Server.TLSConfig == nil {
		err = srv.Server.ListenAndServe()
	} else {
		err = srv.Server.ListenAndServeTLS("server.crt", "server.key")
	}
	if err != nil && err != http.ErrServerClosed {
		log.Fatalf("http-server listen and serve error: %v\n", err)
	}
}

func (srv *HTTP) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), srv.shutdownTimeout)
	defer cancel()
	log.Println("Shutting down http-server")

	return srv.Server.Shutdown(ctx)
}
