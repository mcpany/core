module github.com/mcpany/core

go 1.24.0

toolchain go1.24.11

replace github.com/mcpany/core => ./

require (
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.27.3
	google.golang.org/genproto/googleapis/api v0.0.0-20251111163417-95abcf5c77ba
	google.golang.org/grpc v1.75.1
	google.golang.org/protobuf v1.36.10
)

require (
	golang.org/x/net v0.41.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/text v0.29.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251103181224-f26f9409b101 // indirect
)
