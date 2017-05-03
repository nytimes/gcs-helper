package main

import (
	"io"
	"log"
	"net"
	"net/http"
	"strconv"
	"time"

	"cloud.google.com/go/storage"
	"golang.org/x/net/context"
)

func getHandler(c Config) (http.HandlerFunc, error) {
	client, err := storage.NewClient(context.Background())
	if err != nil {
		return nil, err
	}
	bucketHandle := client.Bucket(c.BucketName)
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		if r.Method != "GET" && r.Method != "HEAD" {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		objectHandle := bucketHandle.Object(r.URL.Path)
		switch r.Method {
		case "HEAD":
			writeHeader(ctx, objectHandle, w)
		case "GET":
			handleGet(ctx, objectHandle, w)
		}
	}, nil
}

func writeHeader(ctx context.Context, object *storage.ObjectHandle, w http.ResponseWriter) error {
	attrs, err := object.Attrs(ctx)
	if err != nil {
		return handleObjectError(err, w)
	}
	w.Header().Set("Cache-Control", attrs.CacheControl)
	w.Header().Set("Content-Type", attrs.ContentType)
	w.Header().Set("Content-Length", strconv.FormatInt(attrs.Size, 10))
	w.Header().Set("Date", time.Now().Format(time.RFC1123))
	w.Header().Set("Last-Modified", attrs.Updated.Format(time.RFC1123))
	w.WriteHeader(http.StatusOK)
	return nil
}

func handleGet(ctx context.Context, object *storage.ObjectHandle, w http.ResponseWriter) error {
	reader, err := object.NewReader(ctx)
	if err != nil {
		return handleObjectError(err, w)
	}
	err = writeHeader(ctx, object, w)
	if err != nil {
		return err
	}
	n, err := io.Copy(w, reader)
	if err != nil {
		log.Printf("failed to download object from S3: %s", err)
		return nil
	}
	if n != reader.Size() {
		log.Printf("wrong number of bytes served from GCS. GCS reports %d bytes, but only %d bytes were transferred", reader.Size(), n)
	}
	return nil
}

func handleObjectError(err error, w http.ResponseWriter) error {
	switch err {
	case storage.ErrBucketNotExist, storage.ErrObjectNotExist:
		http.Error(w, err.Error(), http.StatusNotFound)
	default:
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	return err
}

func main() {
	config, err := loadConfig()
	if err != nil {
		log.Fatal(err)
	}
	handler, err := getHandler(config)
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
