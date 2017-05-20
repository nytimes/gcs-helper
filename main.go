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
	client, err := storage.NewClient(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	handler, err := getHandler(config, client)
	if err != nil {
		log.Fatal(err)
	}
	listener, err := net.Listen("tcp", config.Listen)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Listening on %s...", listener.Addr())
	err = http.Serve(listener, handler)
	if err != nil {
		log.Fatal(err)
	}
}
