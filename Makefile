BINARY_NAME=xkcd

default: build

build:
	go build -o ${BINARY_NAME} ./cmd/xkcd/main.go

test:
	go test ./...