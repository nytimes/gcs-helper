#!/bin/bash -e

export CGO_ENABLED=0

if [ "${TRAVIS_PULL_REQUEST}" != "false" ]; then
	echo >&2 "Skipping image build on pull requests..."
	exit 0
fi

if [ "${TRAVIS_GO_VERSION}" != "1.x" ]; then
	echo >&2 "Skipping docker build on Go ${TRAVIS_GO_VERSION}"
	exit 0
fi

if [ "${TRAVIS_BRANCH}" != "master" ] && [ -z "${TRAVIS_TAG}" ]; then
	echo >&2 "Skipping docker build on branch ${TRAVIS_BRANCH}"
	exit 0
fi

IMAGE_NAME=nytimes/gcs-helper

go build -o gcs-helper
docker login -u "${DOCKER_USERNAME}" -p "${DOCKER_PASSWORD}"
docker build -t ${IMAGE_NAME}:latest .

if [ -n "${TRAVIS_TAG}" ]; then
	docker tag ${IMAGE_NAME}:latest ${IMAGE_NAME}:${TRAVIS_TAG}
	docker tag ${IMAGE_NAME}:latest ${IMAGE_NAME}:${TRAVIS_TAG%%.*}
fi

docker push ${IMAGE_NAME}
