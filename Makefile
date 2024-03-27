BINARY_NAME=myapp

default: build

build:
	go build -o ${BINARY_NAME} ./cmd/cli/main.go

run: build
	./${BINARY_NAME}

test:
	go test ./...