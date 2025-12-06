# syntax=docker/dockerfile:1
FROM golang:1.25.4-alpine AS builder

ENV PROJECT_DIR=/app \
    GO111MODULE=on \
    CGO_ENABLED=0
ENV REDIS_HOST=redis
ENV REDIS_PORT=6379

WORKDIR /app

COPY . .

RUN go get github.com/githubnemo/CompileDaemon
RUN go install github.com/githubnemo/CompileDaemon

RUN go mod download

ENTRYPOINT CompileDaemon -build="go build -o /build/market ./cmd/market" -command="/build/market"
