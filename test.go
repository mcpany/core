package main

import (
	"fmt"
	configv1 "github.com/mcpany/core/proto/config/v1"
)

func main() {
    svc := &configv1.UpstreamServiceConfig{}
    switch svc.WhichServiceConfig() {
    case configv1.UpstreamServiceConfig_McpService_case:
        fmt.Println("McpService", len(svc.GetMcpService().GetTools()))
    default:
        fmt.Println("Unknown")
    }
}
