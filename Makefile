.PHONY: build build-alpine clean test help default

BIN_NAME=githubTop

IT_COMMIT=$(shell git rev-parse HEAD)
GIT_DIRTY=$(shell test -n "`git status --porcelain`" && echo "+CHANGES" || true)
IMAGE_NAME := "marcosquesada/github-top"

default: test

help:
	@echo 'Management commands for githubTop:'
	@echo
	@echo 'Usage:'
	@echo '    make build           Compile the project.'
	@echo '    make get-deps        runs mod install, mostly used for ci.'
	@echo '    make package         Build final docker image with just the go binary inside'
	@echo '    make push            Push tagged images to registry'
	@echo '    make clean           Clean the directory tree.'
	@echo '    make test            Run tests on a compiled project.'
	@echo

build:
	@echo "building ${BIN_NAME}"
	@echo "GOPATH=${GOPATH}"
	go build -o bin/${BIN_NAME} cmd/http/main.go

get-deps:
	go mod vendor

package: get-deps
	@echo "building image ${BIN_NAME} $(GIT_COMMIT)"
	docker build -t marcosquesada/github-top .

push:
	@echo "Pushing docker image to registry: latest $(GIT_COMMIT)"
	docker push marcosquesada/github-top

docker-run:
	@echo "Pulling docker container and run!"
	docker run --rm -it -p 8000:8000 marcosquesada/github-top

run:
	@echo "Run Http App"
	go run main.go http

test:
	go test -v --race ./...

