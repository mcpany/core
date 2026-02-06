package command

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/prompt"
	"github.com/mcpany/core/server/pkg/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestUpstream_Shutdown(t *testing.T) {
	u := NewUpstream()
	err := u.Shutdown(context.Background())
	assert.NoError(t, err)
}

func TestUpstream_Register_DisabledItems(t *testing.T) {
	u := NewUpstream()
	tm := newMockToolManager()
	prm := prompt.NewManager()
	rm := resource.NewManager()

	cmdService := configv1.CommandLineUpstreamService_builder{
		Tools: []*configv1.ToolDefinition{
			configv1.ToolDefinition_builder{Name: proto.String("tool1"), Disable: proto.Bool(true)}.Build(),
		},
		Prompts: []*configv1.PromptDefinition{
			configv1.PromptDefinition_builder{Name: proto.String("prompt1"), Disable: proto.Bool(true)}.Build(),
		},
		Resources: []*configv1.ResourceDefinition{
			configv1.ResourceDefinition_builder{Name: proto.String("resource1"), Disable: proto.Bool(true)}.Build(),
		},
	}.Build()

	config := configv1.UpstreamServiceConfig_builder{
		Id:                 proto.String("disabled-items-service"),
		Name:               proto.String("disabled-items-service"),
		CommandLineService: cmdService,
	}.Build()

	_, tools, _, err := u.Register(context.Background(), config, tm, prm, rm, false)
	require.NoError(t, err)
	assert.Empty(t, tools)
	assert.Empty(t, tm.ListTools())
	assert.Empty(t, prm.ListPrompts())
	assert.Empty(t, rm.ListResources())
}

func TestUpstream_Register_DynamicResourceErrors(t *testing.T) {
	u := NewUpstream()
	tm := newMockToolManager()
	prm := prompt.NewManager()
	rm := resource.NewManager()

	cmdService := configv1.CommandLineUpstreamService_builder{
		Calls: map[string]*configv1.CommandLineCallDefinition{
			"c1": configv1.CommandLineCallDefinition_builder{Id: proto.String("c1")}.Build(),
		},
		Resources: []*configv1.ResourceDefinition{
			configv1.ResourceDefinition_builder{
				Name: proto.String("res1"),
				Dynamic: configv1.DynamicResource_builder{
					CommandLineCall: configv1.CommandLineCallDefinition_builder{Id: proto.String("c1")}.Build(),
				}.Build(),
			}.Build(),
		},
	}.Build()

	config := configv1.UpstreamServiceConfig_builder{
		Id:                 proto.String("dynamic-errors"),
		Name:               proto.String("dynamic-errors"),
		CommandLineService: cmdService,
	}.Build()

	// Register should succeed but resource not added because toolName not found for c1
	_, _, _, err := u.Register(context.Background(), config, tm, prm, rm, false)
	require.NoError(t, err)
	assert.Empty(t, rm.ListResources())
}

func TestUpstream_Register_DisabledPrompt(_ *testing.T) {
	// Covered in DisabledItems
}
