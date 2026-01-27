// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/require"
)

func TestStripSecretsFromService_HunterCoverage(t *testing.T) {
	// Cover empty functions: stripSecretsFromGrpcService, stripSecretsFromOpenapiService
	// And stripSecretsFromMcpCall (inside stripSecretsFromMcpService)

	// gRPC
	svc := &configv1.UpstreamServiceConfig{}
	grpcSvc := &configv1.GrpcUpstreamService{}
	grpcCall := &configv1.GrpcCallDefinition{}
	grpcCall.SetMethod("SomeMethod")
	grpcSvc.SetCalls(map[string]*configv1.GrpcCallDefinition{"call1": grpcCall})
	svc.SetGrpcService(grpcSvc)
	StripSecretsFromService(svc)

	// OpenAPI
	svc = &configv1.UpstreamServiceConfig{}
	openapiSvc := &configv1.OpenapiUpstreamService{}
	openapiCall := &configv1.OpenAPICallDefinition{}
	openapiCall.SetId("op1")
	openapiSvc.SetCalls(map[string]*configv1.OpenAPICallDefinition{"call1": openapiCall})
	svc.SetOpenapiService(openapiSvc)
	StripSecretsFromService(svc)

	// MCP
	svc = &configv1.UpstreamServiceConfig{}
	mcpSvc := &configv1.McpUpstreamService{}
	mcpCall := &configv1.MCPCallDefinition{}
	mcpCall.SetId("tool1")
	mcpSvc.SetCalls(map[string]*configv1.MCPCallDefinition{"call1": mcpCall})

	stdioConn := &configv1.McpStdioConnection{}

	// Secret Value
	sv := &configv1.SecretValue{}
	sv.SetPlainText("secret")

	stdioConn.SetEnv(map[string]*configv1.SecretValue{"KEY": sv})
	mcpSvc.SetStdioConnection(stdioConn)

	svc.SetMcpService(mcpSvc)

	StripSecretsFromService(svc)

    // Check if plain text was stripped
    val := svc.GetMcpService().GetStdioConnection().GetEnv()["KEY"].GetPlainText()
	require.Empty(t, val, "Expected secret to be stripped")

    // Cover HydrateSecretsInService nil/empty checks
    HydrateSecretsInService(nil, nil)
    HydrateSecretsInService(&configv1.UpstreamServiceConfig{}, nil)

	// Filesystem S3
	svc = &configv1.UpstreamServiceConfig{}
	s3Fs := &configv1.S3Fs{}
	s3Fs.SetSecretAccessKey("secret")
	s3Fs.SetSessionToken("token")
	fsSvc := &configv1.FilesystemUpstreamService{}
	fsSvc.SetS3(s3Fs)
	svc.SetFilesystemService(fsSvc)
	StripSecretsFromService(svc)
	require.Empty(t, svc.GetFilesystemService().GetS3().GetSecretAccessKey(), "S3 secret not stripped")

	// Filesystem SFTP
	svc = &configv1.UpstreamServiceConfig{}
	sftpFs := &configv1.SftpFs{}
	sftpFs.SetPassword("pass")
	fsSvc = &configv1.FilesystemUpstreamService{}
	fsSvc.SetSftp(sftpFs)
	svc.SetFilesystemService(fsSvc)
	StripSecretsFromService(svc)
	require.Empty(t, svc.GetFilesystemService().GetSftp().GetPassword(), "SFTP password not stripped")

	// Vector Pinecone
	svc = &configv1.UpstreamServiceConfig{}
	pinecone := &configv1.PineconeVectorDB{}
	pinecone.SetApiKey("key")
	vecSvc := &configv1.VectorUpstreamService{}
	vecSvc.SetPinecone(pinecone)
	svc.SetVectorService(vecSvc)
	StripSecretsFromService(svc)
	require.Empty(t, svc.GetVectorService().GetPinecone().GetApiKey(), "Pinecone API key not stripped")

	// Vector Milvus
	svc = &configv1.UpstreamServiceConfig{}
	milvus := &configv1.MilvusVectorDB{}
	milvus.SetApiKey("key")
	milvus.SetPassword("pass")
	vecSvc = &configv1.VectorUpstreamService{}
	vecSvc.SetMilvus(milvus)
	svc.SetVectorService(vecSvc)
	StripSecretsFromService(svc)
	require.Empty(t, svc.GetVectorService().GetMilvus().GetApiKey(), "Milvus API key not stripped")
	require.Empty(t, svc.GetVectorService().GetMilvus().GetPassword(), "Milvus password not stripped")
}
