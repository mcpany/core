module github.com/mcpany/core/pkg/client

go 1.24.0

require (
	github.com/mcpany/core/server v0.0.0-00010101000000-000000000000
	github.com/gorilla/websocket v1.5.3
	github.com/modelcontextprotocol/go-sdk v1.4.0
	github.com/stretchr/testify v1.11.1
	google.golang.org/grpc v1.79.1
)

replace github.com/mcpany/core/server => ../../server
