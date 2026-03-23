// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitCmd(t *testing.T) {
	tempDir := t.TempDir()
	origDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		_ = os.Chdir(origDir)
	}()

	err = os.Chdir(tempDir)
	require.NoError(t, err)

	cmd := newInitCmd()
	var out bytes.Buffer
	var in bytes.Buffer

	in.WriteString("test-svc\nhttps://test.com\n")

	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetIn(&in)

	err = cmd.Execute()
	require.NoError(t, err)

	output := out.String()
	assert.Contains(t, output, "Successfully generated config.yaml!")

	configPath := "config.yaml"
	assert.FileExists(t, configPath)

	content, err := os.ReadFile(configPath)
	require.NoError(t, err)
	assert.Contains(t, string(content), "test-svc")
	assert.Contains(t, string(content), "https://test.com")
}

func TestInitCmd_Defaults(t *testing.T) {
	tempDir := t.TempDir()
	origDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		_ = os.Chdir(origDir)
	}()

	err = os.Chdir(tempDir)
	require.NoError(t, err)

	cmd := newInitCmd()
	var out bytes.Buffer
	var in bytes.Buffer

	in.WriteString("\n\n")

	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetIn(&in)

	err = cmd.Execute()
	require.NoError(t, err)

	content, err := os.ReadFile("config.yaml")
	require.NoError(t, err)
	assert.Contains(t, string(content), "name: my-service")
	assert.Contains(t, string(content), "address: https://api.example.com")
}
