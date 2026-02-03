export const MCP_OPTIONS_PROTO = `// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

edition = "2023";

package mcpany.mcp_options.v1;

option go_package = "github.com/mcpany/core/proto/mcp_options/v1";

import "google/protobuf/descriptor.proto";

// Option to specify the MCP tool name, applied to a method.
extend google.protobuf.MethodOptions {
   string tool_name = 301009001;
}

// Option to specify the MCP tool description, applied to a method.
extend google.protobuf.MethodOptions {
   string tool_description = 301009002;
}

// Option to specify the MCP prompt name, applied to a method.
extend google.protobuf.MethodOptions {
   string prompt_name = 301009003;
}

// Option to specify the MCP prompt description, applied to a method.
extend google.protobuf.MethodOptions {
   string prompt_description = 301009004;
}

// Option to specify the MCP prompt template, applied to a method.
extend google.protobuf.MethodOptions {
   string prompt_template = 301009005;
}

// Option to specify the MCP resource name, applied to a message.
extend google.protobuf.MessageOptions {
   string resource_name = 301009006;
}

// Option to specify the MCP resource description, applied to a message.
extend google.protobuf.MessageOptions {
     string resource_description = 301009007;
}

// Option to specify the MCP field description, applied to a field.
extend google.protobuf.FieldOptions {
     string field_description = 301009008;
}

// Option to specify the MCP tool read-only hint, applied to a method.
extend google.protobuf.MethodOptions {
   bool mcp_tool_readonly_hint = 301009009;
}

// Option to specify the MCP tool destructive hint, applied to a method.
extend google.protobuf.MethodOptions {
   bool mcp_tool_destructive_hint = 301009010;
}

// Option to specify the MCP tool idempotent hint, applied to a method.
extend google.protobuf.MethodOptions {
   bool mcp_tool_idempotent_hint = 301009011;
}

// Option to specify the MCP tool open-world hint, applied to a method.
extend google.protobuf.MethodOptions {
   bool mcp_tool_openworld_hint = 301009012;
}
`;

export const USER_SERVICE_PROTO = `// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

syntax = "proto3";

package examples.userservice.v1;

// Import the MCP options. The path must be findable by protoc.
// When protoc is run for this file, an include path to the main project's 'proto' dir will be needed.
import "proto/mcp_options/v1/mcp_options.proto";

option go_package = "github.com/mcpany/core/proto/examples/userservice/v1";

// --- Messages ---
message EchoRequest {
    string message = 1 [(mcpany.mcp_options.v1.field_description) = "The message to be echoed."];
}

message EchoResponse {
    string echoed_message = 1 [(mcpany.mcp_options.v1.field_description) = "The echoed message."];
}

message GetDetailsRequest {
    string item_id = 1 [(mcpany.mcp_options.v1.field_description) = "The ID of the item to fetch details for."];
}

message GetDetailsResponse {
    option (mcpany.mcp_options.v1.resource_name) = "ItemDetail";
    option (mcpany.mcp_options.v1.resource_description) = "Represents detailed information about an item.";

    string item_id = 1 [(mcpany.mcp_options.v1.field_description) = "The ID of the item."];
    string detail = 2 [(mcpany.mcp_options.v1.field_description) = "The detailed information about the item."];
    map<string, string> attributes = 3 [(mcpany.mcp_options.v1.field_description) = "A map of attributes for the item."];
}

// --- Service Definition ---
service EchoService {
    // This method will be an MCP tool
    rpc Echo(EchoRequest) returns (EchoResponse) {
        option (mcpany.mcp_options.v1.tool_name) = "EchoTool";
        option (mcpany.mcp_options.v1.tool_description) = "Echoes back the input message. Useful for testing.";
    }

    // This method will also be an MCP tool
    rpc GetDetails(GetDetailsRequest) returns (GetDetailsResponse) {
        option (mcpany.mcp_options.v1.tool_name) = "ItemDetailFetcher";
        option (mcpany.mcp_options.v1.tool_description) = "Fetches details for a given item ID.";
    }
}
`;
