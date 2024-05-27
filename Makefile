default: all

all: build run

build:
	docker-compose build

run:
	docker-compose up -d

test:
	@set CGO_ENABLED=1&& go test -race -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html

e2e-test:
	go test -tags=e2e -v ./e2e

lint:
	golangci-lint run -v

sec: sec-govulncheck sec-trivy

sec-govulncheck:
	govulncheck ./...

sec-trivy:
	trivy fs .
