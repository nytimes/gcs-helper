package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

type codeWrapper struct {
	code int
	http.ResponseWriter
}

func (w *codeWrapper) WriteHeader(code int) {
	w.code = code
	w.ResponseWriter.WriteHeader(code)
}

func getProxyHandler(c Config, client *storage.Client) http.HandlerFunc {
	bucketHandle := client.Bucket(c.BucketName)
	logger := c.logger()
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		defer r.Body.Close()
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		objectName := strings.TrimLeft(r.URL.Path, "/")
		if objectName == "" {
			w.WriteHeader(http.StatusOK)
			return
		}
		resp := codeWrapper{ResponseWriter: w}
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		objectHandle := bucketHandle.Object(objectName)
		switch r.Method {
		case http.MethodHead:
			writeHeader(ctx, objectHandle, &resp, nil, http.StatusOK)
		case http.MethodGet:
			handleGet(ctx, objectHandle, &resp, r)
		}

		// Don't process headers if it's not going to log the request.
		if logger.Level <= logrus.DebugLevel {
			fields := logrus.Fields{
				"method":      r.Method,
				"ellapsed":    time.Since(start).String(),
				"url":         r.URL.RequestURI(),
				"proxyPrefix": c.ProxyPrefix,
				"response":    resp.code,
			}
			for _, header := range c.ProxyLogHeaders {
				if value := r.Header.Get(header); value != "" {
					fields["ReqHeader/"+header] = value
				}
			}
			logger.WithFields(fields).Debug("finished handling request")
		}
	}
}

func writeHeader(ctx context.Context, object *storage.ObjectHandle, w http.ResponseWriter, extra http.Header, status int) error {
	attrs, err := object.Attrs(ctx)
	if err != nil {
		return handleObjectError(err, w)
	}
	if attrs.CacheControl != "" {
		w.Header().Set("Cache-Control", attrs.CacheControl)
	}
	w.Header().Set("Accept-Ranges", "bytes")
	w.Header().Set("Content-Length", strconv.FormatInt(attrs.Size, 10))
	w.Header().Set("Content-Type", attrs.ContentType)
	w.Header().Set("Date", time.Now().Format(time.RFC1123))
	w.Header().Set("Last-Modified", attrs.Updated.Format(time.RFC1123))
	for name, value := range extra {
		w.Header().Set(name, value[0])
	}
	w.WriteHeader(status)
	return nil
}

func handleGet(ctx context.Context, object *storage.ObjectHandle, w http.ResponseWriter, r *http.Request) error {
	offset, end, length := getRange(r)
	reader, err := object.NewRangeReader(ctx, offset, length)
	if err != nil {
		return handleObjectError(err, w)
	}
	extraHeaders := make(http.Header)
	extraHeaders.Set("Content-Length", strconv.FormatInt(reader.Remain(), 10))
	extraHeaders.Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", offset, end, reader.Size()))
	status := http.StatusPartialContent
	if length == -1 {
		status = http.StatusOK
	}
	err = writeHeader(ctx, object, w, extraHeaders, status)
	if err != nil {
		return err
	}
	_, err = io.Copy(w, reader)
	if err != nil {
		log.Printf("failed to download object from GCS: %s", err)
	}
	return nil
}

func getRange(r *http.Request) (offset, end, length int64) {
	length = -1
	if rangeHeader := r.Header.Get("Range"); rangeHeader != "" {
		parts := strings.SplitN(rangeHeader, "=", 2)
		if len(parts) == 2 {
			rangeSpec := strings.SplitN(parts[1], "-", 2)
			if len(rangeSpec) == 2 {
				if rangeStart, err := strconv.ParseInt(rangeSpec[0], 10, 64); err == nil {
					offset = rangeStart
				}
				if rangeEnd, err := strconv.ParseInt(rangeSpec[1], 10, 64); err == nil {
					end = rangeEnd
					length = end - offset + 1
				}
			}
		}
	}
	return offset, end, length
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
