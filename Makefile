BINARY_NAME=xkcd-server

default: server

server:
	go build -o ${BINARY_NAME} ./cmd/xkcdserver/main.go

test:
	go test ./...