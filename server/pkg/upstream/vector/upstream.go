package vector

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/prompt"
	"github.com/mcpany/core/server/pkg/resource"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/mcpany/core/server/pkg/upstream"
	"github.com/mcpany/core/server/pkg/util"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

// ClientFactory is a function that creates a VectorClient.
type ClientFactory func(config *configv1.VectorUpstreamService) (Client, error)

// Upstream implements the upstream.Upstream interface for vector database services.
type Upstream struct {
	clientFactory ClientFactory
}

// NewUpstream creates a new instance of VectorUpstream.
//
// Returns the result.
func NewUpstream() upstream.Upstream {
	return &Upstream{
		clientFactory: defaultClientFactory,
	}
}

func defaultClientFactory(config *configv1.VectorUpstreamService) (Client, error) {
	if t := config.GetPinecone(); t != nil {
		return NewPineconeClient(t)
	}
	if t := config.GetMilvus(); t != nil {
		return NewMilvusClient(t)
	}
	return nil, fmt.Errorf("unsupported vector database type")
}

// Shutdown implements the upstream.Upstream interface.
//
// _ is an unused parameter.
//
// Returns an error if the operation fails.
func (u *Upstream) Shutdown(_ context.Context) error {
	return nil
}

// Register processes the configuration for a vector service.
//
// _ is an unused parameter.
// serviceConfig is the serviceConfig.
// toolManager is the toolManager.
// _ is an unused parameter.
// _ is an unused parameter.
// _ is an unused parameter.
//
// Returns the result.
// Returns the result.
// Returns the result.
// Returns an error if the operation fails.
func (u *Upstream) Register(
	_ context.Context,
	serviceConfig *configv1.UpstreamServiceConfig,
	toolManager tool.ManagerInterface,
	_ prompt.ManagerInterface,
	_ resource.ManagerInterface,
	_ bool,
) (string, []*configv1.ToolDefinition, []*configv1.ResourceDefinition, error) {
	log := logging.GetLogger()

	// Calculate SHA256 for the ID
	h := sha256.New()
	h.Write([]byte(serviceConfig.GetName()))
	serviceConfig.SetId(hex.EncodeToString(h.Sum(nil)))

	// Sanitize the service name
	sanitizedName, err := util.SanitizeServiceName(serviceConfig.GetName())
	if err != nil {
		return "", nil, nil, err
	}
	serviceConfig.SetSanitizedName(sanitizedName)
	serviceID := sanitizedName

	vectorService := serviceConfig.GetVectorService()
	if vectorService == nil {
		return "", nil, nil, fmt.Errorf("vector service config is nil")
	}

	client, err := u.clientFactory(vectorService)
	if err != nil {
		return "", nil, nil, fmt.Errorf("failed to create vector client: %w", err)
	}

	info := &tool.ServiceInfo{
		Name:   serviceConfig.GetName(),
		Config: serviceConfig,
	}
	toolManager.AddServiceInfo(serviceID, info)

	// Define built-in tools
	tools := u.getTools(client)

	discoveredTools := make([]*configv1.ToolDefinition, 0)

	for _, t := range tools {
		toolName := t.Name

		inputSchema, err := structpb.NewStruct(map[string]interface{}{
			"type":       "object",
			"properties": t.Input,
		})
		if err != nil {
			log.Error("Failed to create input schema", "tool", toolName, "error", err)
			continue
		}

		outputSchema, err := structpb.NewStruct(map[string]interface{}{
			"type":       "object",
			"properties": t.Output,
		})
		if err != nil {
			log.Error("Failed to create output schema", "tool", toolName, "error", err)
			continue
		}

		toolDef := configv1.ToolDefinition_builder{
			Name:        proto.String(toolName),
			ServiceId:   proto.String(serviceID),
			Description: proto.String(t.Description),
		}.Build()

		handler := t.Handler
		callable := &vectorCallable{handler: handler}

		// Create a callable tool
		callableTool, err := tool.NewCallableTool(toolDef, serviceConfig, callable, inputSchema, outputSchema)
		if err != nil {
			log.Error("Failed to create callable tool", "tool", toolName, "error", err)
			continue
		}

		if err := toolManager.AddTool(callableTool); err != nil {
			log.Error("Failed to add tool", "tool", toolName, "error", err)
			continue
		}

		discoveredTools = append(discoveredTools, toolDef)
	}

	log.Info("Registered vector service", "serviceID", serviceID, "tools", len(discoveredTools))
	return serviceID, discoveredTools, nil, nil
}

type vectorCallable struct {
	handler func(ctx context.Context, args map[string]interface{}) (map[string]interface{}, error)
}

// Call executes the vector tool with the given arguments.
// It accepts a context and an execution request containing arguments,
// and returns the result of the tool execution or an error.
func (c *vectorCallable) Call(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
	return c.handler(ctx, req.Arguments)
}

type vectorToolDef struct {
	Name        string
	Description string
	Input       map[string]interface{}
	Output      map[string]interface{}
	Handler     func(ctx context.Context, args map[string]interface{}) (map[string]interface{}, error)
}

// Client interface for different vector DB implementations.
type Client interface {
	// Query searches for the nearest vectors in the database.
	// It accepts a context, a query vector, the number of results to return (topK),
	// a metadata filter, and a namespace.
	// It returns a map containing the search results or an error.
	Query(ctx context.Context, vector []float32, topK int64, filter map[string]interface{}, namespace string) (map[string]interface{}, error)

	// Upsert inserts or updates vectors in the database.
	// It accepts a context, a list of vectors (each as a map), and a namespace.
	// It returns a map containing the operation result (e.g., upserted count) or an error.
	Upsert(ctx context.Context, vectors []map[string]interface{}, namespace string) (map[string]interface{}, error)

	// Delete removes vectors from the database.
	// It accepts a context, a list of IDs to delete, a namespace, and an optional metadata filter.
	// It returns a map containing the operation result or an error.
	Delete(ctx context.Context, ids []string, namespace string, filter map[string]interface{}) (map[string]interface{}, error)

	// DescribeIndexStats retrieves statistics about the vector index.
	// It accepts a context and an optional metadata filter.
	// It returns a map containing the index statistics or an error.
	DescribeIndexStats(ctx context.Context, filter map[string]interface{}) (map[string]interface{}, error)
}

func (u *Upstream) getTools(client Client) []vectorToolDef {
	return []vectorToolDef{
		{
			Name:        "query_vectors",
			Description: "Query the vector database for similar vectors.",
			Input: map[string]interface{}{
				"vector": map[string]interface{}{
					"type":        "array",
					"items":       map[string]interface{}{"type": "number"},
					"description": "The query vector.",
				},
				"top_k":     map[string]interface{}{"type": "integer", "description": "Number of results to return."},
				"filter":    map[string]interface{}{"type": "object", "description": "Metadata filter."},
				"namespace": map[string]interface{}{"type": "string", "description": "Namespace to query."},
			},
			Output: map[string]interface{}{
				"matches": map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "object"}},
			},
			Handler: func(ctx context.Context, args map[string]interface{}) (map[string]interface{}, error) {
				vectorInterface, ok := args["vector"].([]interface{})
				if !ok {
					return nil, fmt.Errorf("vector is required and must be an array")
				}
				vector := make([]float32, len(vectorInterface))
				for i, v := range vectorInterface {
					if f, ok := v.(float64); ok {
						vector[i] = float32(f)
					} else {
						return nil, fmt.Errorf("vector elements must be numbers")
					}
				}

				topK, ok := args["top_k"].(float64)
				if !ok {
					topK = 10
				}

				var filter map[string]interface{}
				if f, ok := args["filter"].(map[string]interface{}); ok {
					filter = f
				}

				namespace, _ := args["namespace"].(string)

				return client.Query(ctx, vector, int64(topK), filter, namespace)
			},
		},
		{
			Name:        "upsert_vectors",
			Description: "Upsert vectors into the database.",
			Input: map[string]interface{}{
				"vectors": map[string]interface{}{
					"type": "array",
					"items": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"id":       map[string]interface{}{"type": "string"},
							"values":   map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "number"}},
							"metadata": map[string]interface{}{"type": "object"},
						},
						"required": []interface{}{"id", "values"},
					},
					"description": "List of vectors to upsert.",
				},
				"namespace": map[string]interface{}{"type": "string", "description": "Namespace to upsert into."},
			},
			Output: map[string]interface{}{
				"upserted_count": map[string]interface{}{"type": "integer"},
			},
			Handler: func(ctx context.Context, args map[string]interface{}) (map[string]interface{}, error) {
				vectorsInterface, ok := args["vectors"].([]interface{})
				if !ok {
					return nil, fmt.Errorf("vectors is required")
				}

				vectors := make([]map[string]interface{}, len(vectorsInterface))
				for i, v := range vectorsInterface {
					if vm, ok := v.(map[string]interface{}); ok {
						vectors[i] = vm
					} else {
						return nil, fmt.Errorf("vectors must be objects")
					}
				}

				namespace, _ := args["namespace"].(string)
				return client.Upsert(ctx, vectors, namespace)
			},
		},
		{
			Name:        "delete_vectors",
			Description: "Delete vectors from the database.",
			Input: map[string]interface{}{
				"ids": map[string]interface{}{
					"type":        "array",
					"items":       map[string]interface{}{"type": "string"},
					"description": "List of IDs to delete.",
				},
				"namespace": map[string]interface{}{"type": "string", "description": "Namespace to delete from."},
				"filter":    map[string]interface{}{"type": "object", "description": "Metadata filter (optional, if IDs not provided)."},
				"deleteAll": map[string]interface{}{"type": "boolean", "description": "Delete all vectors in namespace."},
			},
			Output: map[string]interface{}{
				"success": map[string]interface{}{"type": "boolean"},
			},
			Handler: func(ctx context.Context, args map[string]interface{}) (map[string]interface{}, error) {
				var ids []string
				if idsInterface, ok := args["ids"].([]interface{}); ok {
					ids = make([]string, len(idsInterface))
					for i, v := range idsInterface {
						ids[i] = fmt.Sprint(v)
					}
				}

				namespace, _ := args["namespace"].(string)

				var filter map[string]interface{}
				if f, ok := args["filter"].(map[string]interface{}); ok {
					filter = f
				}

				return client.Delete(ctx, ids, namespace, filter)
			},
		},
		{
			Name:        "describe_index_stats",
			Description: "Get statistics about the index.",
			Input: map[string]interface{}{
				"filter": map[string]interface{}{"type": "object", "description": "Filter stats by metadata."},
			},
			Output: map[string]interface{}{
				"namespaces":       map[string]interface{}{"type": "object"},
				"dimension":        map[string]interface{}{"type": "integer"},
				"indexFullness":    map[string]interface{}{"type": "number"},
				"totalVectorCount": map[string]interface{}{"type": "integer"},
			},
			Handler: func(ctx context.Context, args map[string]interface{}) (map[string]interface{}, error) {
				var filter map[string]interface{}
				if f, ok := args["filter"].(map[string]interface{}); ok {
					filter = f
				}
				return client.DescribeIndexStats(ctx, filter)
			},
		},
	}
}
