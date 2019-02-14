FROM    golang:1.12rc1-alpine AS build
ENV     CGO_ENABLED 0
RUN     apk add --no-cache git
ADD     . /code
WORKDIR /code
RUN     go install

FROM alpine:3.9
RUN  apk add --no-cache ca-certificates
COPY --from=build /go/bin/gcs-helper /usr/bin/gcs-helper
ENTRYPOINT ["/usr/bin/gcs-helper"]
