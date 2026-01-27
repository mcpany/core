// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package schemaconv

import (
	"testing"

	userservicev1 "github.com/mcpany/core/proto/examples/userservice/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFieldDescriptionExtraction(t *testing.T) {
	// Get the file descriptor for userservice.proto
	fileDesc := userservicev1.File_proto_examples_userservice_v1_userservice_proto
	require.NotNil(t, fileDesc)

	// Find EchoRequest message
	msgDesc := fileDesc.Messages().ByName("EchoRequest")
	require.NotNil(t, msgDesc)

	// Find the 'message' field which has the annotation
	// string message = 1 [(mcpany.mcp_options.v1.field_description) = "The message to be echoed."];
	fieldDesc := msgDesc.Fields().ByName("message")
	require.NotNil(t, fieldDesc)

	// Call fieldToSchema directly as we are in the same package
	schema, err := fieldToSchema(fieldDesc, 0)
	require.NoError(t, err)

	// Verify description
	desc, ok := schema["description"]
	require.True(t, ok, "description field not found in schema")
	assert.Equal(t, "The message to be echoed.", desc)
}
