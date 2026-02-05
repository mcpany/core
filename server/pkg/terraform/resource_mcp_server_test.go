// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package terraform

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTerraformResource(t *testing.T) {
	schema := Schema()
	assert.Contains(t, schema, "name")
	assert.Contains(t, schema, "port")

	err := Create(&ResourceMCPServer{Name: "test", Port: 9090})
	assert.NoError(t, err)

	res, err := Read("test")
	assert.NoError(t, err)
	assert.Equal(t, "test", res.Name)
}
