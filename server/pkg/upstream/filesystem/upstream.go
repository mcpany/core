// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package filesystem provides the filesystem upstream implementation.
package filesystem

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"sync"

	"github.com/alexliesenfeld/health"
	configv1 "github.com/mcpany/core/proto/config/v1"
	mcphealth "github.com/mcpany/core/server/pkg/health"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/prompt"
	"github.com/mcpany/core/server/pkg/resource"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/mcpany/core/server/pkg/upstream"
	"github.com/mcpany/core/server/pkg/upstream/filesystem/provider"
	"github.com/mcpany/core/server/pkg/util"

	jsoniter "github.com/json-iterator/go"
	"github.com/spf13/afero"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

var fastJSON = jsoniter.ConfigCompatibleWithStandardLibrary

// Upstream implements the upstream.Upstream interface for filesystem services.
type Upstream struct {
	mu      sync.Mutex
	closers []io.Closer
	checker health.Checker
}

// NewUpstream creates a new instance of FilesystemUpstream.
//
// Returns the result.
func NewUpstream() upstream.Upstream {
	return &Upstream{
		closers: make([]io.Closer, 0),
	}
}

// Shutdown implements the upstream.Upstream interface.
//
// _ is an unused parameter.
//
// Returns an error if the operation fails.
func (u *Upstream) Shutdown(_ context.Context) error {
	u.mu.Lock()
	defer u.mu.Unlock()
	if u.checker != nil {
		u.checker.Stop()
	}
	for _, c := range u.closers {
		_ = c.Close()
	}
	u.closers = nil
	return nil
}

// Register processes the configuration for a filesystem service.
//
// ctx is the context for the request.
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
	ctx context.Context,
	serviceConfig *configv1.UpstreamServiceConfig,
	toolManager tool.ManagerInterface,
	_ prompt.ManagerInterface,
	_ resource.ManagerInterface,
	_ bool,
) (string, []*configv1.ToolDefinition, []*configv1.ResourceDefinition, error) {
	log := logging.GetLogger()

	// Calculate SHA256 for the ID if not set
	if serviceConfig.GetId() == "" {
		h := sha256.New()
		h.Write([]byte(serviceConfig.GetName()))
		serviceConfig.SetId(hex.EncodeToString(h.Sum(nil)))
	}

	// Sanitize the service name
	sanitizedName, err := util.SanitizeServiceName(serviceConfig.GetName())
	if err != nil {
		return "", nil, nil, err
	}
	serviceConfig.SetSanitizedName(sanitizedName)
	serviceID := sanitizedName

	fsService := serviceConfig.GetFilesystemService()
	if fsService == nil {
		return "", nil, nil, fmt.Errorf("filesystem service config is nil")
	}

	// Create the filesystem provider
	prov, err := u.createProvider(ctx, fsService)
	if err != nil {
		return "", nil, nil, fmt.Errorf("failed to create filesystem provider: %w", err)
	}

	// Register closer for the provider
	u.mu.Lock()
	u.closers = append(u.closers, prov)
	u.mu.Unlock()

	fs := prov.GetFs()

	// Initialize and start health checker
	if u.checker != nil {
		u.checker.Stop()
	}
	u.checker = mcphealth.NewChecker(serviceConfig)
	if u.checker != nil {
		u.checker.Start()
	}

	info := &tool.ServiceInfo{
		Name:   serviceConfig.GetName(),
		Config: serviceConfig,
	}
	toolManager.AddServiceInfo(serviceID, info)

	// Define built-in tools
	tools := u.getSupportedTools(fsService, prov, fs)

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
			Name:      proto.String(toolName),
			ServiceId: proto.String(serviceID),
		}.Build()

		handler := t.Handler
		callable := &fsCallable{handler: handler}

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

	log.Info("Registered filesystem service", "serviceID", serviceID, "tools", len(discoveredTools))
	return serviceID, discoveredTools, nil, nil
}

type fsCallable struct {
	handler func(ctx context.Context, args map[string]interface{}) (map[string]interface{}, error)
}

// Call executes the filesystem tool with the provided request arguments.
// It returns the result of the tool execution or an error.
func (c *fsCallable) Call(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
	args := req.Arguments
	if args == nil && len(req.ToolInputs) > 0 {
		if err := fastJSON.Unmarshal(req.ToolInputs, &args); err != nil {
			return nil, fmt.Errorf("failed to unmarshal arguments: %w", err)
		}
	}
	return c.handler(ctx, args)
}

func (u *Upstream) createProvider(ctx context.Context, config *configv1.FilesystemUpstreamService) (provider.Provider, error) {
	var prov provider.Provider
	var err error

	// Determine the backend filesystem
	if config.GetTmpfs() != nil {
		prov = provider.NewTmpfsProvider()
	} else if config.GetHttp() != nil {
		return nil, fmt.Errorf("http filesystem is not yet supported")
	} else if zip := config.GetZip(); zip != nil {
		prov, err = provider.NewZipProvider(zip)
		if err != nil {
			return nil, err
		}
	} else if gcs := config.GetGcs(); gcs != nil {
		prov, err = provider.NewGcsProvider(ctx, gcs)
		if err != nil {
			return nil, err
		}
	} else if sftp := config.GetSftp(); sftp != nil {
		prov, err = provider.NewSftpProvider(sftp)
		if err != nil {
			return nil, err
		}
	} else if s3 := config.GetS3(); s3 != nil {
		prov, err = provider.NewS3Provider(s3)
		if err != nil {
			return nil, err
		}
	} else if os := config.GetOs(); os != nil {
		prov = provider.NewLocalProvider(os, config.GetRootPaths(), config.GetAllowedPaths(), config.GetDeniedPaths(), config.GetSymlinkMode())
	} else {
		// Fallback to OsFs for backward compatibility if root_paths is set?
		// Or defaulting to OsFs.
		// Use nil for OsFs config, effectively default.
		prov = provider.NewLocalProvider(nil, config.GetRootPaths(), config.GetAllowedPaths(), config.GetDeniedPaths(), config.GetSymlinkMode())
	}

	// Wrap with ReadOnly if requested.
	// Since ReadOnly is a property of the service config, we might want to wrap the Fs returned by provider.
	// However, provider interface returns Fs. We can wrap it here?
	// But the provider methods (ResolvePath) are separate.
	// We can wrap the provider?
	// For now, let's wrap the Fs in getSupportedTools usage or create a wrapper provider.
	// Actually, the simplest way is to handle ReadOnly check in the tools themselves as it was.
	// But wait, createFilesystem returned a ReadOnlyFs.
	// If we change the fs returned by provider, we modify the provider's state?
	// No, provider.GetFs() returns a stored fs. We can't easily change it without casting.
	// So we should probably handle ReadOnly in the tools (write_file, delete_file).
	// The original implementation returned `afero.NewReadOnlyFs(baseFs)`.
	// Let's defer this check to the tools.

	return prov, nil
}

func (u *Upstream) getSupportedTools(fsService *configv1.FilesystemUpstreamService, prov provider.Provider, fs afero.Fs) []filesystemToolDef {
	// Wrap fs with ReadOnly if requested
	if fsService.GetReadOnly() {
		fs = afero.NewReadOnlyFs(fs)
	}

	return getTools(prov, fs, fsService.GetReadOnly(), fsService.GetRootPaths())
}
