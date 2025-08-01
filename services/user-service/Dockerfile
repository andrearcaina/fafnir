FROM golang:1.24-alpine

RUN apk add --no-cache git && \
    go install github.com/air-verse/air@latest

WORKDIR /app/services/user-service

COPY ../shared/go.mod /app/shared/go.mod
COPY ../services/user-service/go.mod .
COPY ../services/user-service/go.sum .

RUN go mod download

CMD ["air"]