package openapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/prompt"
	"github.com/mcpany/core/server/pkg/resource"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/proto"
)

func TestOpenAPIUpstream_Register_Content(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockToolManager := tool.NewMockManagerInterface(ctrl)
	mockPromptManager := prompt.NewMockManagerInterface(ctrl)
	mockResourceManager := resource.NewMockManagerInterface(ctrl)

	u := NewOpenAPIUpstream()

	specContent := `
openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /test:
    get:
      operationId: getTest
      responses:
        '200':
          description: OK
`

	config := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("test-service"),
		OpenapiService: configv1.OpenapiUpstreamService_builder{
			SpecContent: proto.String(specContent),
			Address:     proto.String("http://127.0.0.1"),
		}.Build(),
	}.Build()

	// Expectations
	mockToolManager.EXPECT().AddServiceInfo(gomock.Any(), gomock.Any()).Do(func(id string, info *tool.ServiceInfo) {
		assert.Contains(t, id, "test-service")
		assert.Equal(t, "test-service", info.Name)
	})

	// We expect AddTool to be called for "getTest"
	mockToolManager.EXPECT().GetTool("getTest").Return(nil, false) // Check for existence
	mockToolManager.EXPECT().AddTool(gomock.Any()).Return(nil)

	// Register
	serviceID, tools, resources, err := u.Register(context.Background(), config, mockToolManager, mockPromptManager, mockResourceManager, false)

	assert.NoError(t, err)
	assert.NotEmpty(t, serviceID)
	assert.Len(t, tools, 1)
	assert.Empty(t, resources)
}

func TestOpenAPIUpstream_Register_URL(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockToolManager := tool.NewMockManagerInterface(ctrl)
	mockPromptManager := prompt.NewMockManagerInterface(ctrl)
	mockResourceManager := resource.NewMockManagerInterface(ctrl)

	specContent := `
openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /remote:
    get:
      operationId: getRemote
      responses:
        '200':
          description: OK
`
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(specContent))
	}))
	defer ts.Close()

	u := NewOpenAPIUpstream()

	config := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("remote-service"),
		OpenapiService: configv1.OpenapiUpstreamService_builder{
			SpecUrl: proto.String(ts.URL),
		}.Build(),
	}.Build()

	mockToolManager.EXPECT().AddServiceInfo(gomock.Any(), gomock.Any())
	mockToolManager.EXPECT().GetTool("getRemote").Return(nil, false)
	mockToolManager.EXPECT().AddTool(gomock.Any()).Return(nil)

	serviceID, tools, _, err := u.Register(context.Background(), config, mockToolManager, mockPromptManager, mockResourceManager, false)

	assert.NoError(t, err)
	assert.NotEmpty(t, serviceID)
	assert.Len(t, tools, 1)
}

func TestOpenAPIUpstream_Register_WithPromptsAndResources(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockToolManager := tool.NewMockManagerInterface(ctrl)
	mockPromptManager := prompt.NewMockManagerInterface(ctrl)
	mockResourceManager := resource.NewMockManagerInterface(ctrl)

	u := NewOpenAPIUpstream()

	specContent := `
openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /test:
    get:
      operationId: getTest
      responses:
        '200':
          description: OK
`

	config := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("test-service"),
		OpenapiService: configv1.OpenapiUpstreamService_builder{
			SpecContent: proto.String(specContent),
			Address:     proto.String("http://127.0.0.1"),
			Prompts: []*configv1.PromptDefinition{
				configv1.PromptDefinition_builder{
					Name: proto.String("test-prompt"),
					Messages: []*configv1.PromptMessage{
						configv1.PromptMessage_builder{
							Role: configv1.PromptMessage_USER.Enum(),
							Text: configv1.TextContent_builder{
								Text: proto.String("test template"),
							}.Build(),
						}.Build(),
					},
				}.Build(),
			},
			Resources: []*configv1.ResourceDefinition{
				configv1.ResourceDefinition_builder{
					Name: proto.String("test-resource"),
					Dynamic: configv1.DynamicResource_builder{
						HttpCall: configv1.HttpCallDefinition_builder{
							Id: proto.String("getTest"),
						}.Build(),
					}.Build(),
				}.Build(),
			},
		}.Build(),
	}.Build()

	// Service Info
	mockToolManager.EXPECT().AddServiceInfo(gomock.Any(), gomock.Any()).Do(func(_ string, info *tool.ServiceInfo) {
		assert.Equal(t, "test-service", info.Name)
	})

	// Tool Registration
	mockToolManager.EXPECT().GetTool("getTest").Return(nil, false)
	var registeredTool tool.Tool
	mockToolManager.EXPECT().AddTool(gomock.Any()).DoAndReturn(func(t tool.Tool) error {
		registeredTool = t
		return nil
	})

	// Prompt Registration
	mockPromptManager.EXPECT().AddPrompt(gomock.Any()).Do(func(p prompt.Prompt) {
		assert.Equal(t, "test-service.test-prompt", p.Prompt().Name)
	})

	// Resource Registration
	mockToolManager.EXPECT().GetTool("test-service.getTest").Return(nil, false).DoAndReturn(func(_ string) (tool.Tool, bool) {
		return registeredTool, true
	})

	mockResourceManager.EXPECT().AddResource(gomock.Any()).Do(func(r resource.Resource) {
		assert.Equal(t, "test-resource", r.Resource().Name)
	})

	// Register
	serviceID, tools, _, err := u.Register(context.Background(), config, mockToolManager, mockPromptManager, mockResourceManager, false)

	assert.NoError(t, err)
	assert.NotEmpty(t, serviceID)
	assert.Len(t, tools, 1)
}
