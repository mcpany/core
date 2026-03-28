package main

import (
	"fmt"
	"github.com/mcpany/core/server/pkg/app"
)

func main() {
	fmt.Printf("Templates: %d\n", len(app.BuiltinServiceTemplates))
}
