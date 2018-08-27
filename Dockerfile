FROM linuxkit/ca-certificates:v0.4 AS ca-certificates

FROM    golang:1.11 AS build
ENV     CGO_ENABLED 0
ADD     . /code
WORKDIR /code
RUN     go test
RUN     go install

FROM alpine:3.8
COPY --from=build /go/bin/gcs-helper /usr/bin/gcs-helper
COPY --from=ca-certificates / /
ENTRYPOINT ["/usr/bin/gcs-helper"]
