// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestFilesystemHealthCheck(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "mcpany-fs-health-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	existingDir := filepath.Join(tempDir, "existing")
	err = os.Mkdir(existingDir, 0755)
	require.NoError(t, err)

	missingDir := filepath.Join(tempDir, "missing")

	app := NewApplication()

	tests := []struct {
		name           string
		services       []*configv1.UpstreamServiceConfig
		expectedStatus string
		expectInMsg    string
	}{
		{
			name:           "No filesystem services",
			services:       []*configv1.UpstreamServiceConfig{},
			expectedStatus: "ok",
		},
		{
			name: "Valid filesystem service",
			services: []*configv1.UpstreamServiceConfig{
				{
					Name: proto.String("svc-1"),
					ServiceConfig: &configv1.UpstreamServiceConfig_FilesystemService{
						FilesystemService: &configv1.FilesystemUpstreamService{
							RootPaths: map[string]string{
								"/data": existingDir,
							},
						},
					},
				},
			},
			expectedStatus: "ok",
		},
		{
			name: "Inaccessible root path",
			services: []*configv1.UpstreamServiceConfig{
				{
					Name: proto.String("svc-bad"),
					ServiceConfig: &configv1.UpstreamServiceConfig_FilesystemService{
						FilesystemService: &configv1.FilesystemUpstreamService{
							RootPaths: map[string]string{
								"/data": missingDir,
							},
						},
					},
				},
			},
			expectedStatus: "degraded",
			expectInMsg:    "inaccessible",
		},
		{
			name: "Path is a file, not a directory",
			services: []*configv1.UpstreamServiceConfig{
				{
					Name: proto.String("svc-file"),
					ServiceConfig: &configv1.UpstreamServiceConfig_FilesystemService{
						FilesystemService: &configv1.FilesystemUpstreamService{
							RootPaths: map[string]string{
								"/data": filepath.Join(existingDir, "somefile.txt"),
							},
						},
					},
				},
			},
			expectedStatus: "degraded",
			expectInMsg:    "not a directory",
		},
	}

	// Create a file for the last test case
	err = os.WriteFile(filepath.Join(existingDir, "somefile.txt"), []byte("data"), 0644)
	require.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app.ServiceRegistry = &TestMockServiceRegistry{services: tt.services}
			res := app.filesystemHealthCheck(context.Background())
			assert.Equal(t, tt.expectedStatus, res.Status)
			if tt.expectInMsg != "" {
				assert.Contains(t, res.Message, tt.expectInMsg)
			}
		})
	}
}
