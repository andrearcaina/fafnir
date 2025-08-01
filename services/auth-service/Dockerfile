FROM golang:1.24-alpine

RUN apk add --no-cache git && \
    go install github.com/air-verse/air@latest

WORKDIR /app/services/auth-service

COPY ../shared/go.mod /app/shared/go.mod
COPY ../services/auth-service/go.mod .
COPY ../services/auth-service/go.sum .

RUN go mod download

CMD ["air"]