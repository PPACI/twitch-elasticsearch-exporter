FROM golang:1.16 as builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./
RUN go build -o twitch-surveillance .

FROM debian:buster-slim
RUN apt update && apt install -yq ca-certificates

COPY --from=builder /app/twitch-surveillance /usr/local/bin/twitch-surveillace

ENTRYPOINT ["twitch-sql-exporter"]