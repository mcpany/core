// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package graphql provides GraphQL upstream integration.
package graphql

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/machinebox/graphql"
	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/prompt"
	"github.com/mcpany/core/server/pkg/resource"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/mcpany/core/server/pkg/upstream"
	configv1 "github.com/mcpany/core/proto/config/v1"
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

type graphqlUpstream struct{}

// NewGraphQLUpstream creates a new GraphQL upstream.
func NewGraphQLUpstream() upstream.Upstream {
	return &graphqlUpstream{}
}

// Shutdown shuts down the upstream.
func (g *graphqlUpstream) Shutdown(_ context.Context) error {
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

// Callable implements the Callable interface for GraphQL queries.
type Callable struct {
	client        *graphql.Client
	query         string
	authenticator auth.UpstreamAuthenticator
	address       string
}

// Call executes the GraphQL query.
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
func (g *graphqlUpstream) Register(
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

	authenticator, err := auth.NewUpstreamAuthenticator(serviceConfig.GetUpstreamAuthentication())
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
						Type struct {
							Name   string
							Kind   string
							OfType *struct {
								Name string
								Kind string
							}
						}
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
					typeName := arg.Type.Name
					if arg.Type.Kind == "NON_NULL" {
						typeName = arg.Type.OfType.Name
					}
					inputSchema.Fields[arg.Name] = &structpb.Value{
						Kind: &structpb.Value_StructValue{
							StructValue: &structpb.Struct{
								Fields: map[string]*structpb.Value{
									"type": {
										Kind: &structpb.Value_StringValue{
											StringValue: mapGraphQLTypeToJSONSchemaType(typeName),
										},
									},
								},
							},
						},
					}
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
					typeName := arg.Type.Name
					if arg.Type.Kind == "NON_NULL" {
						typeName = arg.Type.OfType.Name + "!"
					}
					sb.WriteString(fmt.Sprintf("$%s: %s", arg.Name, typeName))
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
