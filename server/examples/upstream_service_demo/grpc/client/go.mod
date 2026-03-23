module github.com/mcpany/core/upstream_service/grpc/client

go 1.24.0

require (
	github.com/mcpany/core/upstream_service/grpc/greeter_server v0.0.0
	google.golang.org/grpc v1.77.0
)

require (
	go.opentelemetry.io/otel/metric v1.39.0 // indirect
	go.opentelemetry.io/otel/sdk v1.39.0 // indirect
	go.opentelemetry.io/otel/trace v1.39.0 // indirect
	golang.org/x/net v0.48.0 // indirect
	golang.org/x/sys v0.39.0 // indirect
	golang.org/x/text v0.32.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251202230838-ff82c1b0f217 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
)

replace github.com/mcpany/core/upstream_service/grpc/greeter_server => ../greeter_server
