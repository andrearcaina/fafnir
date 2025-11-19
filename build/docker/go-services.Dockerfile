ARG GO_VERSION=1.24.5

# Development stage
FROM golang:${GO_VERSION}-alpine AS development
ARG SERVICE_NAME

RUN apk add --no-cache git curl && \
    go install github.com/air-verse/air@v1.62.0

WORKDIR /app

COPY src/shared/ ./src/shared/
COPY src/${SERVICE_NAME}/ ./src/${SERVICE_NAME}/
COPY tools/.air.toml .air.toml

RUN cd src/shared && go mod tidy
RUN cd src/${SERVICE_NAME} && go mod tidy

EXPOSE 8080

CMD ["air", "-c", ".air.toml"]

# Builder stage
FROM golang:${GO_VERSION}-alpine AS builder
ARG SERVICE_NAME

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

COPY src/shared/ ./src/shared/
COPY src/${SERVICE_NAME}/ ./src/${SERVICE_NAME}/

RUN cd src/shared && go mod tidy
RUN cd src/${SERVICE_NAME} && go mod tidy

RUN cd src/${SERVICE_NAME} && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o /app/main ./cmd/server

# Production stage
FROM scratch AS production
ARG SERVICE_NAME

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /app/main /main

USER 65534:65534

EXPOSE 8080

ENTRYPOINT ["/main"]