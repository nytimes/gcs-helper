FROM golang:1.8-alpine AS build
ENV  CGO_ENABLED 0
ADD  . /go/src/github.com/NYTimes/gcs-helper
RUN  go install github.com/NYTimes/gcs-helper

FROM alpine:3.5
COPY --from=build /go/bin/gcs-helper /usr/bin/gcs-helper
ENTRYPOINT ["/usr/bin/gcs-helper"]
