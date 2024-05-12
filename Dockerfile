FROM golang:1.22 as BuildStage

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /xkcdserver cmd/xkcdserver/main.go

EXPOSE 8080

FROM alpine:latest

WORKDIR /
COPY --from=BuildStage /xkcdserver /xkcdserver
COPY --from=BuildStage app/config.yaml /config.yaml
COPY --from=BuildStage app/extended_stopwords_eng.txt /extended_stopwords_eng.txt

CMD ["/xkcdserver"]