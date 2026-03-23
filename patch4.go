package main

import (
	"fmt"
	"os"
	"strings"
)

func main() {
	content, err := os.ReadFile("server/tests/integration/e2e_helpers.go")
	if err != nil {
		panic(err)
	}

	str := string(content)

	// Fix the unused callDef error 1
	str = strings.Replace(str, `	callID := "call-" + toolDef.GetName()
	callDef := configv1.HttpCallDefinition_builder{
		Id:           &callID,
		EndpointPath: &endpointPath,
		Method:       &method,
	}.Build()

	tool := &configv1.Tool{
		Tool: &configv1.Tool_Definition{
			Definition: toolDef,
		},
	}`, `	tool := &configv1.Tool{
		Tool: &configv1.Tool_Definition{
			Definition: toolDef,
		},
	}`, 1)

	// Fix the unused callDef error 2
	str = strings.Replace(str, `	toolDef := configv1.ToolDefinition_builder{Name: &operationID}.Build()
	callID := "call-" + toolDef.GetName()
	callDef := configv1.HttpCallDefinition_builder{
		Id:           &callID,
		EndpointPath: &endpointPath,
		Method:       &method,
	}.Build()

	tool := &configv1.Tool{
		Tool: &configv1.Tool_Definition{
			Definition: toolDef,
		},
	}`, `	toolDef := configv1.ToolDefinition_builder{Name: &operationID}.Build()

	tool := &configv1.Tool{
		Tool: &configv1.Tool_Definition{
			Definition: toolDef,
		},
	}`, 1)

	err = os.WriteFile("server/tests/integration/e2e_helpers.go", []byte(str), 0644)
	if err != nil {
		panic(err)
	}
	fmt.Println("Patched successfully")
}
