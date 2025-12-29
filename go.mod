module github.com/mcpany/core

go 1.24.0

toolchain go1.24.11

replace github.com/mcpany/core => ./

require (
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.27.3
	github.com/modelcontextprotocol/go-sdk v1.1.0
	google.golang.org/genproto/googleapis/api v0.0.0-20251111163417-95abcf5c77ba
	google.golang.org/grpc v1.75.1
	google.golang.org/protobuf v1.36.10
)

require (
	github.com/fclairamb/afero-s3 v0.3.1
	github.com/aws/aws-sdk-go v1.55.8
	github.com/google/jsonschema-go v0.3.0 // indirect
)
