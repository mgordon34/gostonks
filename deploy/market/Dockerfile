# syntax=docker/dockerfile:1

FROM golang:1.25.4-alpine AS builder
WORKDIR /src

RUN apk add --no-cache git

COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/market ./cmd/market

FROM alpine:latest
WORKDIR /app

RUN apk add --no-cache ca-certificates

COPY --from=builder /out/market ./market

ENV REDIS_HOST=redis
ENV REDIS_PORT=6379

ENTRYPOINT ["/app/market"]
