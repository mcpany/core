// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestToolIntegrityConfig(t *testing.T) {
	toolDef := configv1.ToolDefinition_builder{
		Name:        proto.String("test-tool"),
		Description: proto.String("A test tool"),
		ServiceId:   proto.String("test-service"),
	}.Build()

	// 1. Convert without integrity - should pass VerifyIntegrity
	pbTool, err := ConvertToolDefinitionToProto(toolDef, nil, nil)
	require.NoError(t, err)
	require.NoError(t, VerifyIntegrity(pbTool))

	// 2. Add correct integrity
	// Calculate hash of the tool proto without integrity
	marshaler := proto.MarshalOptions{Deterministic: true}
	data, err := marshaler.Marshal(pbTool)
	require.NoError(t, err)
	hash := sha256.Sum256(data)
	hashStr := hex.EncodeToString(hash[:])

	toolDef.SetIntegrity(configv1.Integrity_builder{
		Hash:      proto.String(hashStr),
		Algorithm: proto.String("sha256"),
	}.Build())

	pbToolWithIntegrity, err := ConvertToolDefinitionToProto(toolDef, nil, nil)
	require.NoError(t, err)
	require.True(t, pbToolWithIntegrity.HasIntegrity())
	require.NoError(t, VerifyIntegrity(pbToolWithIntegrity))

	// 3. Add incorrect integrity
	toolDef.GetIntegrity().SetHash("incorrect-hash")
	pbToolBadIntegrity, err := ConvertToolDefinitionToProto(toolDef, nil, nil)
	require.NoError(t, err)
	err = VerifyIntegrity(pbToolBadIntegrity)
	require.Error(t, err)
	require.Contains(t, err.Error(), "integrity check failed")
}

func TestVerifyConfigIntegrity(t *testing.T) {
	toolDef := configv1.ToolDefinition_builder{
		Name:        proto.String("config-test-tool"),
		Description: proto.String("A config test tool"),
	}.Build()

	// 1. No integrity -> Pass
	require.NoError(t, VerifyConfigIntegrity(toolDef))

	// 2. Correct integrity -> Pass
	// Calculate hash manually
	hashStr, err := CalculateConfigHash(toolDef)
	require.NoError(t, err)

	toolDef.SetIntegrity(configv1.Integrity_builder{
		Hash:      proto.String(hashStr),
		Algorithm: proto.String("sha256"),
	}.Build())

	require.NoError(t, VerifyConfigIntegrity(toolDef))

	// 3. Incorrect integrity -> Fail
	toolDef.GetIntegrity().SetHash("bad-hash")
	err = VerifyConfigIntegrity(toolDef)
	require.Error(t, err)
	require.Contains(t, err.Error(), "integrity check failed")
}
