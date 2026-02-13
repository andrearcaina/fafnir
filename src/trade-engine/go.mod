module fafnir/trade-engine

go 1.24.5

require (
	fafnir/shared v0.0.0-00010101000000-000000000000
	// Dependencies will be added by go mod tidy, but listing core ones here
	github.com/go-chi/chi/v5 v5.2.5
	google.golang.org/grpc v1.78.0
	google.golang.org/protobuf v1.36.11
)

require github.com/nats-io/nats.go v1.48.0

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/klauspost/compress v1.18.2 // indirect
	github.com/nats-io/nkeys v0.4.12 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	github.com/redis/go-redis/v9 v9.17.2 // indirect
	golang.org/x/crypto v0.46.0 // indirect
	golang.org/x/net v0.48.0 // indirect
	golang.org/x/sys v0.39.0 // indirect
	golang.org/x/text v0.32.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251222181119-0a764e51fe1b // indirect
)

replace fafnir/shared => ../shared
