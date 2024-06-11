FROM golang:1.22.4 as BuildStage

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /xkcdserver cmd/xkcdserver/main.go

EXPOSE 8080

FROM alpine:latest

WORKDIR /
COPY --from=BuildStage /xkcdserver /xkcdserver

CMD ["/xkcdserver"]