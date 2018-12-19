package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"github.com/sirupsen/logrus"
)

const maxTry = 5

type codeWrapper struct {
	code int
	http.ResponseWriter
}

func (w *codeWrapper) WriteHeader(code int) {
	w.code = code
	w.ResponseWriter.WriteHeader(code)
}

func getProxyHandler(c Config, client *storage.Client) http.HandlerFunc {
	logger := c.logger()
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		defer r.Body.Close()
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if r.URL.Path == "/" {
			w.WriteHeader(http.StatusOK)
			return
		}
		resp := codeWrapper{ResponseWriter: w}
		ctx, cancel := context.WithTimeout(context.Background(), c.Proxy.Timeout)
		defer cancel()
		obj := objectHandle(&c, client, r)
		var err error

		switch r.Method {
		case http.MethodHead:
			err = writeHeader(ctx, obj, &resp, nil, http.StatusOK)
		case http.MethodGet:
			err = handleGet(ctx, obj, &resp, r)
		}

		if err != nil || logger.Level <= logrus.DebugLevel {
			fields := logrus.Fields{
				"method":        r.Method,
				"ellapsed":      time.Since(start).String(),
				"url":           r.URL.RequestURI(),
				"proxyEndpoint": c.Proxy.Endpoint,
				"response":      resp.code,
			}
			for _, header := range c.Proxy.LogHeaders {
				if value := r.Header.Get(header); value != "" {
					fields["ReqHeader/"+header] = value
				}
			}
			entry := logger.WithFields(fields)
			if err != nil {
				entry.WithError(err).Error("failed to handle request")
			} else {
				entry.Debug("finished handling request")
			}
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
	reader, err := getReader(ctx, object, offset, length, maxTry)
	if err != nil {
		return handleObjectError(err, w)
	}
	defer reader.Close()
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
	return err
}

func getReader(ctx context.Context, object *storage.ObjectHandle, offset, length int64, try int) (*storage.Reader, error) {
	if try == 0 {
		return nil, errors.New("max try exceeded")
	}
	reader, err := object.NewRangeReader(ctx, offset, length)
	switch err {
	case nil, context.DeadlineExceeded, context.Canceled, storage.ErrBucketNotExist, storage.ErrObjectNotExist:
		return reader, err
	default:
		return getReader(ctx, object, offset, length, try-1)
	}
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
		return nil
	default:
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	return err
}

func objectHandle(c *Config, client *storage.Client, r *http.Request) *storage.ObjectHandle {
	bucketName := c.BucketName
	objectName := strings.TrimLeft(r.URL.Path, "/")
	if c.Proxy.BucketOnPath {
		pos := strings.Index(objectName, "/")
		bucketName = objectName[:pos]
		objectName = objectName[pos+1:]
	}
	return client.Bucket(bucketName).Object(objectName)
}
