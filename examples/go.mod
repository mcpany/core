module github.com/mcpxy/core/examples

go 1.24.3

require (
	github.com/gorilla/websocket v1.5.4-0.20250319132907-e064f32e3674
	google.golang.org/grpc v1.75.0
	google.golang.org/protobuf v1.36.8
)

require (
	go.opentelemetry.io/otel v1.38.0 // indirect
	go.opentelemetry.io/otel/sdk/metric v1.38.0 // indirect
	golang.org/x/net v0.43.0 // indirect
	golang.org/x/sys v0.35.0 // indirect
	golang.org/x/text v0.28.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250825161204-c5933d9347a5 // indirect
)

replace github.com/mcpxy/core => ../
