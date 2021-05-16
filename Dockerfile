FROM golang:1.15 as builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY pkg pkg
COPY main.go .
RUN go build -o twitch-surveillance .

FROM debian:buster-slim
RUN apt update && apt install -yq ca-certificates

COPY --from=builder /app/twitch-surveillace /usr/local/bin/twitch-surveillace

ENTRYPOINT ["twitch-sql-exporter"]