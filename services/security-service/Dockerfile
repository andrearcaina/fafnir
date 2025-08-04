FROM golang:1.24-alpine

RUN apk add --no-cache git && \
    go install github.com/air-verse/air@latest

WORKDIR /app/services/security-service

COPY ../shared/go.mod /app/shared/go.mod
COPY ../services/security-service/go.mod .
COPY ../services/security-service/go.sum .

RUN go mod download

CMD ["air"]