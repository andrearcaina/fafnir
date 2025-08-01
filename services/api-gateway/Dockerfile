FROM golang:1.24-alpine

RUN apk add --no-cache git && \
    go install github.com/air-verse/air@latest

WORKDIR /app/services/api-gateway

COPY ../services/api-gateway/go.mod .
COPY ../services/api-gateway/go.sum .

RUN go mod download

CMD ["air"]