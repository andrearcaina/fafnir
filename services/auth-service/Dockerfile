FROM golang:1.24-alpine

RUN apk add --no-cache git && \
    go install github.com/air-verse/air@latest

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

WORKDIR /app/services/auth-service

CMD ["air"]