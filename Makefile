default: all

test:
	go test ./...

all: build run

build:
	docker-compose build

run:
	docker-compose up -d