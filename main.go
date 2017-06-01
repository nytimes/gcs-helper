package main

import (
	"log"
	"net"
	"net/http"

	"cloud.google.com/go/storage"
	"golang.org/x/net/context"
)

func main() {
	config, err := loadConfig()
	if err != nil {
		log.Fatal(err)
	}
	logger := config.logger()
	client, err := storage.NewClient(context.Background())
	if err != nil {
		logger.WithError(err).Fatal("failed to create storage client instance")
	}
	handler, err := getHandler(config, client)
	if err != nil {
		logger.WithError(err).Fatal("failed to get handle instance")
	}
	listener, err := net.Listen("tcp", config.Listen)
	if err != nil {
		logger.WithField("listenAddr", config.Listen).WithError(err).Fatal("failed to start listener")
	}

	logger.Infof("Listening on %s...", listener.Addr())
	err = http.Serve(listener, handler)
	if err != nil {
		logger.WithError(err).Fatal("failed to start server")
	}
}
