// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package grpc

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/pool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/types/descriptorpb"
)

func TestFindMethodDescriptor_EdgeCases(t *testing.T) {
	// Create a minimal FileDescriptorSet
	fds := &descriptorpb.FileDescriptorSet{
		File: []*descriptorpb.FileDescriptorProto{
			{
				Name:    proto.String("test.proto"),
				Package: proto.String("pkg"),
				MessageType: []*descriptorpb.DescriptorProto{
					{Name: proto.String("Request")},
					{Name: proto.String("Response")},
				},
				Service: []*descriptorpb.ServiceDescriptorProto{
					{
						Name: proto.String("Service"),
						Method: []*descriptorpb.MethodDescriptorProto{
							{
								Name:       proto.String("Method"),
								InputType:  proto.String(".pkg.Request"),
								OutputType: proto.String(".pkg.Response"),
							},
						},
					},
				},
			},
		},
	}
	files, err := protodesc.NewFiles(fds)
	require.NoError(t, err)

	// Valid case
	md, err := findMethodDescriptor(files, "pkg.Service/Method")
	assert.NoError(t, err)
	assert.NotNil(t, md)

	// Invalid format (no separator)
	_, err = findMethodDescriptor(files, "pkgServiceMethod")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid full method name")

	// Service not found
	_, err = findMethodDescriptor(files, "pkg.Unknown/Method")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "could not find descriptor")

	// Method not found
	_, err = findMethodDescriptor(files, "pkg.Service/Unknown")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "method 'Unknown' not found")
}

func TestGRPCUpstream_Register_ProtoParseError(t *testing.T) {
	poolManager := pool.NewManager()
	upstream := NewUpstream(poolManager)

	config := &configv1.UpstreamServiceConfig{
		Name: proto.String("proto-fail-service"),
		ServiceConfig: &configv1.UpstreamServiceConfig_GrpcService{
			GrpcService: &configv1.GrpcUpstreamService{
				Address:       proto.String("127.0.0.1:50051"),
				UseReflection: proto.Bool(false),
				ProtoDefinitions: []*configv1.ProtoDefinition{
					{
						ProtoRef: &configv1.ProtoDefinition_ProtoFile{
							ProtoFile: &configv1.ProtoFile{
								FileName: proto.String("test.proto"),
								FileRef: &configv1.ProtoFile_FileContent{
									FileContent: `syntax = "proto3"; invalid syntax`,
								},
							},
						},
					},
				},
			},
		},
	}

	_, _, _, err := upstream.Register(context.Background(), config, NewMockToolManager(), nil, nil, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse proto definitions")
}
