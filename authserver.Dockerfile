FROM golang:1.22.4 as BuildStage

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /authserver cmd/authserver/main.go

EXPOSE 8080

FROM alpine:latest

WORKDIR /
COPY --from=BuildStage /authserver /authserver
COPY --from=BuildStage app/config/authserver.yaml config/authserver.yaml

CMD ["/authserver"]