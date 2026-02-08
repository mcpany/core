// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package graphql provides GraphQL upstream integration.
package graphql

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/machinebox/graphql"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/prompt"
	"github.com/mcpany/core/server/pkg/resource"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/mcpany/core/server/pkg/upstream"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

const introspectionQuery = `
  query IntrospectionQuery {
    __schema {
      queryType { name }
      mutationType { name }
      types {
        ...FullType
      }
    }
  }

  fragment FullType on __Type {
    kind
    name
    description
    fields(includeDeprecated: true) {
      name
      description
      args {
        ...InputValue
      }
      type {
        ...TypeRef
      }
    }
    inputFields {
      ...InputValue
    }
    interfaces {
      ...TypeRef
    }
    enumValues(includeDeprecated: true) {
      name
      description
    }
    possibleTypes {
      ...TypeRef
    }
  }

  fragment InputValue on __InputValue {
    name
    description
    type { ...TypeRef }
    defaultValue
  }

  fragment TypeRef on __Type {
    kind
    name
    ofType {
      kind
      name
      ofType {
        kind
        name
        ofType {
          kind
          name
        }
      }
    }
  }
`

// Upstream implements the upstream.Upstream interface for GraphQL services.
//
// Summary: implements the upstream.Upstream interface for GraphQL services.
type Upstream struct{}

// NewGraphQLUpstream creates a new GraphQL upstream.
//
// Summary: creates a new GraphQL upstream.
//
// Parameters:
//   None.
//
// Returns:
//   - upstream.Upstream: The upstream.Upstream.
func NewGraphQLUpstream() upstream.Upstream {
	return &Upstream{}
}

// Shutdown shuts down the upstream.
//
// Summary: shuts down the upstream.
//
// Parameters:
//   - _: context.Context. The _.
//
// Returns:
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
func (g *Upstream) Shutdown(_ context.Context) error {
	return nil
}

func mapGraphQLTypeToJSONSchemaType(typeName string) string {
	switch typeName {
	case "String", "ID":
		return "string"
	case "Int", "Float":
		return "number"
	case "Boolean":
		return "boolean"
	default:
		return "object"
	}
}

type graphQLType struct {
	Kind   string       `json:"kind"`
	Name   *string      `json:"name"`
	OfType *graphQLType `json:"ofType"`
}

func convertGraphQLTypeToJSONSchema(t *graphQLType) *structpb.Value {
	if t == nil {
		return &structpb.Value{Kind: &structpb.Value_StructValue{StructValue: &structpb.Struct{Fields: map[string]*structpb.Value{"type": {Kind: &structpb.Value_StringValue{StringValue: "object"}}}}}}
	}

	switch t.Kind {
	case "NON_NULL":
		return convertGraphQLTypeToJSONSchema(t.OfType)
	case "LIST":
		itemsSchema := convertGraphQLTypeToJSONSchema(t.OfType)
		return &structpb.Value{
			Kind: &structpb.Value_StructValue{
				StructValue: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"type":  {Kind: &structpb.Value_StringValue{StringValue: "array"}},
						"items": itemsSchema,
					},
				},
			},
		}
	default: // SCALAR, OBJECT, etc.
		typeName := ""
		if t.Name != nil {
			typeName = *t.Name
		}
		jsonType := mapGraphQLTypeToJSONSchemaType(typeName)
		return &structpb.Value{
			Kind: &structpb.Value_StructValue{
				StructValue: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"type": {Kind: &structpb.Value_StringValue{StringValue: jsonType}},
					},
				},
			},
		}
	}
}

// Callable implements the Callable interface for GraphQL queries.
//
// Summary: implements the Callable interface for GraphQL queries.
type Callable struct {
	client        *graphql.Client
	query         string
	authenticator auth.UpstreamAuthenticator
	address       string
}

// Call executes the GraphQL query.
//
// Summary: executes the GraphQL query.
//
// Parameters:
//   - ctx: context.Context. The context for the operation.
//   - req: *tool.ExecutionRequest. The req.
//
// Returns:
//   - any: The any.
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
func (c *Callable) Call(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
	graphqlReq := graphql.NewRequest(c.query)
	for key, value := range req.Arguments {
		graphqlReq.Var(key, value)
	}
	if c.authenticator != nil {
		dummyReq, err := http.NewRequestWithContext(ctx, "POST", c.address, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create dummy request: %w", err)
		}
		if err := c.authenticator.Authenticate(dummyReq); err != nil {
			return nil, fmt.Errorf("failed to authenticate graphql query: %w", err)
		}
		graphqlReq.Header = dummyReq.Header
	}
	var respData any
	if err := c.client.Run(ctx, graphqlReq, &respData); err != nil {
		return nil, fmt.Errorf("failed to run graphql query: %w", err)
	}
	return respData, nil
}

// Register inspects the GraphQL upstream service and registers its capabilities.
//
// Summary: inspects the GraphQL upstream service and registers its capabilities.
//
// Parameters:
//   - ctx: context.Context. The context for the operation.
//   - serviceConfig: *configv1.UpstreamServiceConfig. The serviceConfig.
//   - toolManager: tool.ManagerInterface. The toolManager.
//   - _: prompt.ManagerInterface. The _.
//   - _: resource.ManagerInterface. The _.
//   - _: bool. The _.
//
// Returns:
//   - string: The string.
//   - []*configv1.ToolDefinition: The []*configv1.ToolDefinition.
//   - []*configv1.ResourceDefinition: The []*configv1.ResourceDefinition.
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
func (g *Upstream) Register(
	ctx context.Context,
	serviceConfig *configv1.UpstreamServiceConfig,
	toolManager tool.ManagerInterface,
	_ prompt.ManagerInterface,
	_ resource.ManagerInterface,
	_ bool,
) (string, []*configv1.ToolDefinition, []*configv1.ResourceDefinition, error) {
	graphqlConfig := serviceConfig.GetGraphqlService()
	if graphqlConfig == nil {
		return "", nil, nil, fmt.Errorf("missing graphql service config")
	}

	address := graphqlConfig.GetAddress()
	if address == "" {
		return "", nil, nil, fmt.Errorf("graphql service address is required")
	}
	uURL, err := url.ParseRequestURI(address)
	if err != nil {
		return "", nil, nil, fmt.Errorf("invalid graphql service address: %w", err)
	}
	if uURL.Scheme != "http" && uURL.Scheme != "https" {
		return "", nil, nil, fmt.Errorf("invalid graphql service address scheme: %s (must be http or https)", uURL.Scheme)
	}

	authenticator, err := auth.NewUpstreamAuthenticator(serviceConfig.GetUpstreamAuth())
	if err != nil {
		return "", nil, nil, fmt.Errorf("failed to create upstream authenticator: %w", err)
	}

	client := graphql.NewClient(graphqlConfig.GetAddress())

	req := graphql.NewRequest(introspectionQuery)
	if authenticator != nil {
		dummyReq, err := http.NewRequestWithContext(ctx, "POST", graphqlConfig.GetAddress(), nil)
		if err != nil {
			return "", nil, nil, fmt.Errorf("failed to create dummy request: %w", err)
		}
		if err := authenticator.Authenticate(dummyReq); err != nil {
			return "", nil, nil, fmt.Errorf("failed to authenticate introspection query: %w", err)
		}
		req.Header = dummyReq.Header
	}

	var respData struct {
		Schema struct {
			QueryType    struct{ Name string } `json:"queryType"`
			MutationType struct{ Name string } `json:"mutationType"`
			Types        []struct {
				Kind   string
				Name   string
				Fields []struct {
					Name string
					Args []struct {
						Name string
						Type graphQLType `json:"type"`
					} `json:"args"`
					Type struct {
						Kind   string
						Name   string
						Fields []struct {
							Name string
						} `json:"fields"`
					} `json:"type"`
				} `json:"fields"`
			} `json:"types"`
		} `json:"__schema"`
	}

	if err := client.Run(ctx, req, &respData); err != nil {
		return "", nil, nil, fmt.Errorf("failed to run introspection query: %w", err)
	}

	var toolDefs []*configv1.ToolDefinition

	log.Printf("Registering tools for service: %s", serviceConfig.GetName())
	for _, t := range respData.Schema.Types {
		operationType := ""
		switch t.Name {
		case respData.Schema.QueryType.Name:
			operationType = "query"
		case respData.Schema.MutationType.Name:
			operationType = "mutation"
		}

		if operationType != "" {
			for _, field := range t.Fields {
				toolName := fmt.Sprintf("%s-%s", serviceConfig.GetName(), field.Name)
				log.Printf("Creating tool: %s", toolName)
				inputSchema := &structpb.Struct{
					Fields: make(map[string]*structpb.Value),
				}
				for _, arg := range field.Args {
					inputSchema.Fields[arg.Name] = convertGraphQLTypeToJSONSchema(&arg.Type)
				}

				callID := "graphql"
				toolDef := configv1.ToolDefinition_builder{
					Name:        proto.String(toolName),
					Description: proto.String(field.Name),
					CallId:      proto.String(callID),
					ServiceId:   proto.String(serviceConfig.GetName()),
				}.Build()

				toolDefs = append(toolDefs, toolDef)

				var sb strings.Builder
				sb.WriteString(operationType)
				sb.WriteString(" (")
				i := 0
				for _, arg := range field.Args {
					// We need to reconstruct the GraphQL type string for the query variable definition
					// This part also needs to be recursive or handled better if we want full support,
					// but for now let's stick to what was there or slightly improve it.
					// The previous code had a loop/recursion implicit via checks.
					// But wait, constructing the query string requires the Type Name.
					// If it is a LIST, the name is wrapped in [].
					// I need a helper for this too.

					sb.WriteString(fmt.Sprintf("$%s: %s", arg.Name, formatGraphQLType(&arg.Type)))
					if i < len(field.Args)-1 {
						sb.WriteString(", ")
					}
					i++
				}
				sb.WriteString(") { ")
				sb.WriteString(field.Name)
				sb.WriteString("(")
				i = 0
				for _, arg := range field.Args {
					sb.WriteString(fmt.Sprintf("%s: $%s", arg.Name, arg.Name))
					if i < len(field.Args)-1 {
						sb.WriteString(", ")
					}
					i++
				}
				sb.WriteString(") ")
				selectionSet := ""
				if call, ok := graphqlConfig.GetCalls()[field.Name]; ok {
					selectionSet = call.GetSelectionSet()
				}

				if selectionSet != "" {
					sb.WriteString(selectionSet)
				} else if len(field.Type.Fields) > 0 {
					sb.WriteString("{ ")
					for i, f := range field.Type.Fields {
						sb.WriteString(f.Name)
						if i < len(field.Type.Fields)-1 {
							sb.WriteString(" ")
						}
					}
					sb.WriteString(" }")
				}

				sb.WriteString(" }")

				callable := &Callable{client: client, query: sb.String(), authenticator: authenticator, address: graphqlConfig.GetAddress()}

				t, err := tool.NewCallableTool(toolDef, serviceConfig, callable, inputSchema, nil)
				if err != nil {
					return "", nil, nil, fmt.Errorf("failed to create tool %s: %w", toolName, err)
				}
				if err := toolManager.AddTool(t); err != nil {
					return "", nil, nil, fmt.Errorf("failed to add tool %s: %w", toolName, err)
				}
			}
		}
	}

	return serviceConfig.GetName(), toolDefs, nil, nil
}

func formatGraphQLType(t *graphQLType) string {
	if t == nil {
		return ""
	}
	switch t.Kind {
	case "NON_NULL":
		return formatGraphQLType(t.OfType) + "!"
	case "LIST":
		return "[" + formatGraphQLType(t.OfType) + "]"
	default:
		if t.Name != nil {
			return *t.Name
		}
		return ""
	}
}
