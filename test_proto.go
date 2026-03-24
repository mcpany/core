package main

import (
	"context"
	"fmt"
	configv1 "github.com/mcpany/core/proto/config/v1"
)

func main() {
	var t configv1.ToolDefinition
	t.SetName("hello")
	t.SetDisable(true)
	fmt.Println(t.GetName(), t.GetDisable())

    var svc configv1.UpstreamServiceConfig
    svc.SetTools([]*configv1.ToolDefinition{&t})
    fmt.Println(len(svc.GetTools()))
}
