FROM linuxkit/ca-certificates:v0.4 AS ca-certificates

FROM golang:1.10-alpine AS build
ENV  CGO_ENABLED 0
ADD  . /go/src/github.com/NYTimes/gcs-helper
RUN  go test github.com/NYTimes/gcs-helper
RUN  go install github.com/NYTimes/gcs-helper

FROM alpine:3.7
COPY --from=build /go/bin/gcs-helper /usr/bin/gcs-helper
COPY --from=ca-certificates / /
ENTRYPOINT ["/usr/bin/gcs-helper"]
