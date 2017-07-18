package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	"cloud.google.com/go/storage"
	"github.com/google/gops/agent"
	"golang.org/x/net/context"
	"google.golang.org/api/option"
)

const version = "1.5"

func main() {
	err := agent.Listen(&agent.Options{NoShutdownCleanup: true})
	if err != nil {
		log.Fatalf("could not start gops agent: %v", err)
	}
	defer agent.Close()
	handleFlags()
	config, err := loadConfig()
	if err != nil {
		log.Fatal(err)
	}
	logger := config.logger()
	hc := http.Client{
		Timeout: config.ClientConfig.Timeout,
		Transport: &http.Transport{
			IdleConnTimeout: config.ClientConfig.IdleConnTimeout,
			MaxIdleConns:    config.ClientConfig.MaxIdleConns,
		},
	}
	client, err := storage.NewClient(context.Background(), option.WithHTTPClient(&hc))
	if err != nil {
		logger.WithError(err).Fatal("failed to create storage client instance")
	}
	handler := getHandler(config, client)
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

func handleFlags() {
	printVersion := flag.Bool("version", false, "print version and exit")
	flag.Parse()
	if *printVersion {
		fmt.Printf("gcs-helper %s\n", version)
		os.Exit(0)
	}
}
