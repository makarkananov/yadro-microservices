FROM golang:1.22.4 as BuildStage

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /webserver cmd/webserver/main.go

EXPOSE 8081

FROM alpine:latest

WORKDIR /
COPY --from=BuildStage /webserver /webserver
COPY --from=BuildStage app/internal/adapter/handler/web/templates/ /templates

CMD ["/webserver"]