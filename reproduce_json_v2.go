package main

import (
	"fmt"
	"log"

	configpb "github.com/mcpany/core/proto/config/v1"
	"google.golang.org/protobuf/encoding/protojson"
)

func main() {
	// JSON from test-data.ts
	jsonStr := `
{
    "command_line_service": {
        "command": "echo",
        "communication_protocol": "COMMUNICATION_PROTOCOL_JSON",
        "tools": [
            {
                "name": "get_complex_data",
                "description": "Returns complex data for UI testing",
                "call_id": "get_complex_data",
                "input_schema": {
                    "type": "object",
                    "properties": {
                        "dummy": { "type": "string" }
                    }
                }
            }
        ]
    }
}
`
	config := &configpb.UpstreamServiceConfig{}
	err := protojson.Unmarshal([]byte(jsonStr), config)
	if err != nil {
		log.Fatalf("Unmarshal failed: %v", err)
	}

	clService := config.GetCommandLineService()
	if clService == nil {
		log.Fatal("CommandLineService is nil")
	}

	for _, tool := range clService.GetTools() {
		fmt.Printf("Tool: %s\n", tool.GetName())
		fmt.Printf("InputSchema: %v\n", tool.GetInputSchema())
		
		schema := tool.GetInputSchema()
		if schema == nil {
			fmt.Println("Schema is nil")
			continue
		}
		
		// Check for "type" field
		typeVal, ok := schema.GetFields()["type"]
		if !ok {
			fmt.Println("Type field missing!")
			// Dump fields to see what's there
			for k, v := range schema.GetFields() {
				fmt.Printf("Field: %s = %v\n", k, v)
			}
		} else {
			fmt.Printf("Type: %v\n", typeVal.GetStringValue())
		}
	}
}
