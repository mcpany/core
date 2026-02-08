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
