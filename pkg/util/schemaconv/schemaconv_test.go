/*
 * Copyright 2025 Author(s) of MCP-XY
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package schemaconv

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	configv1 "github.com/mcpxy/core/proto/config/v1"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/structpb"
)

// Mock types for testing generic functions
type mockConfigParameter struct {
	schema *configv1.ParameterSchema
}

func (m *mockConfigParameter) GetSchema() *configv1.ParameterSchema {
	return m.schema
}

type mockMcpFieldParameter struct {
	name        string
	description string
	typ         string
}

func (m *mockMcpFieldParameter) GetName() string {
	return m.name
}

func (m *mockMcpFieldParameter) GetDescription() string {
	return m.description
}

func (m *mockMcpFieldParameter) GetType() string {
	return m.typ
}

func strPtr(s string) *string {
	return &s
}

func TestConfigSchemaToProtoProperties(t *testing.T) {
	tests := []struct {
		name    string
		params  []*mockConfigParameter
		want    *structpb.Struct
		wantErr bool
	}{
		{
			name: "successful conversion",
			params: []*mockConfigParameter{
				{schema: configv1.ParameterSchema_builder{
					Name:        strPtr("param1"),
					Description: strPtr("desc1"),
					Type:        configv1.ParameterType_STRING.Enum(),
				}.Build()},
				{schema: configv1.ParameterSchema_builder{
					Name:        strPtr("param2"),
					Description: strPtr("desc2"),
					Type:        configv1.ParameterType_INTEGER.Enum(),
				}.Build()},
			},
			want: &structpb.Struct{
				Fields: map[string]*structpb.Value{
					"param1": {
						Kind: &structpb.Value_StructValue{
							StructValue: &structpb.Struct{
								Fields: map[string]*structpb.Value{
									"type":        {Kind: &structpb.Value_StringValue{StringValue: "string"}},
									"description": {Kind: &structpb.Value_StringValue{StringValue: "desc1"}},
								},
							},
						},
					},
					"param2": {
						Kind: &structpb.Value_StructValue{
							StructValue: &structpb.Struct{
								Fields: map[string]*structpb.Value{
									"type":        {Kind: &structpb.Value_StringValue{StringValue: "integer"}},
									"description": {Kind: &structpb.Value_StringValue{StringValue: "desc2"}},
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name:   "empty params",
			params: []*mockConfigParameter{},
			want: &structpb.Struct{
				Fields: map[string]*structpb.Value{},
			},
			wantErr: false,
		},
		{
			name: "nil schema",
			params: []*mockConfigParameter{
				{schema: nil},
			},
			want: &structpb.Struct{
				Fields: map[string]*structpb.Value{},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ConfigSchemaToProtoProperties(tt.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("ConfigSchemaToProtoProperties() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, got, protocmp.Transform()); diff != "" {
				t.Errorf("ConfigSchemaToProtoProperties() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestMcpFieldsToProtoProperties(t *testing.T) {
	tests := []struct {
		name    string
		params  []*mockMcpFieldParameter
		want    *structpb.Struct
		wantErr bool
	}{
		{
			name: "successful conversion",
			params: []*mockMcpFieldParameter{
				{name: "param1", description: "desc1", typ: "TYPE_STRING"},
				{name: "param2", description: "desc2", typ: "TYPE_INT32"},
				{name: "param3", description: "desc3", typ: "TYPE_BOOL"},
				{name: "param4", description: "desc4", typ: "TYPE_DOUBLE"},
			},
			want: &structpb.Struct{
				Fields: map[string]*structpb.Value{
					"param1": {
						Kind: &structpb.Value_StructValue{
							StructValue: &structpb.Struct{
								Fields: map[string]*structpb.Value{
									"type":        {Kind: &structpb.Value_StringValue{StringValue: "string"}},
									"description": {Kind: &structpb.Value_StringValue{StringValue: "desc1"}},
								},
							},
						},
					},
					"param2": {
						Kind: &structpb.Value_StructValue{
							StructValue: &structpb.Struct{
								Fields: map[string]*structpb.Value{
									"type":        {Kind: &structpb.Value_StringValue{StringValue: "integer"}},
									"description": {Kind: &structpb.Value_StringValue{StringValue: "desc2"}},
								},
							},
						},
					},
					"param3": {
						Kind: &structpb.Value_StructValue{
							StructValue: &structpb.Struct{
								Fields: map[string]*structpb.Value{
									"type":        {Kind: &structpb.Value_StringValue{StringValue: "boolean"}},
									"description": {Kind: &structpb.Value_StringValue{StringValue: "desc3"}},
								},
							},
						},
					},
					"param4": {
						Kind: &structpb.Value_StructValue{
							StructValue: &structpb.Struct{
								Fields: map[string]*structpb.Value{
									"type":        {Kind: &structpb.Value_StringValue{StringValue: "number"}},
									"description": {Kind: &structpb.Value_StringValue{StringValue: "desc4"}},
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name:   "empty params",
			params: []*mockMcpFieldParameter{},
			want: &structpb.Struct{
				Fields: map[string]*structpb.Value{},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := McpFieldsToProtoProperties(tt.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("McpFieldsToProtoProperties() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, got, protocmp.Transform()); diff != "" {
				t.Errorf("McpFieldsToProtoProperties() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
