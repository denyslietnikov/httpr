APP := $(shell basename $(shell git remote get-url origin))
REGISTRY := dletnikov
VERSION=$(shell git describe --tags --abbrev=0)-$(shell git rev-parse --short HEAD)
#linux linux windows
TARGETOS=linux
#amd64 arm64
TARGETARCH=amd64

format:
	gofmt -s -w ./src

lint:
	golint

test:
	go test -v

get:
	go get ./src

build: format get
	CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -v -o bin/app src/main.go

image:
	docker build . -t ${REGISTRY}/${APP}:${VERSION}-${TARGETARCH}  --build-arg TARGETARCH=${TARGETARCH}

push:
	docker push ${REGISTRY}/${APP}:${VERSION}-${TARGETARCH}

clean:
	rm -rf bin
	docker rmi ${REGISTRY}/${APP}:${VERSION}-${TARGETARCH}