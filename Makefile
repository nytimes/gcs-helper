SHELL := /bin/sh

BINARY_NAME ?= gcs-helper
DOCKER_REGISTRY ?= mintel
DOCKER_IMAGE = ${DOCKER_REGISTRY}/${BINARY_NAME}

VERSION ?= $(shell echo `git symbolic-ref -q --short HEAD || git describe --tags --exact-match` | tr '[/]' '-')
DOCKER_TAG ?= ${VERSION}

ARTIFACTS = /tmp/artifacts

build : gcs-helper
.PHONY : build

gcs-helper : main.go
	@echo "building go binary"
	@CGO_ENABLED=0 GOOS=linux go build .
