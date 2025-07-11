FROM golang:1.24-alpine

RUN apk add --no-cache git && \
    go install github.com/air-verse/air@latest

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY services/auth-service ./services/auth-service
COPY shared ./shared

WORKDIR /app/services/auth-service

CMD ["air"]
