// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package browser

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/prompt"
	"github.com/mcpany/core/server/pkg/resource"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/mcpany/core/server/pkg/upstream"
	"github.com/mcpany/core/server/pkg/util"
)

// Upstream implements the upstream.Upstream interface for browser services.
type Upstream struct {
	mu       sync.Mutex
	services []*Service
}

// NewUpstream creates a new instance of BrowserUpstream.
func NewUpstream() upstream.Upstream {
	return &Upstream{
		services: make([]*Service, 0),
	}
}

// Shutdown implements the upstream.Upstream interface.
func (u *Upstream) Shutdown(_ context.Context) error {
	u.mu.Lock()
	defer u.mu.Unlock()
	for _, s := range u.services {
		s.Close()
	}
	u.services = nil
	return nil
}

// Register processes the configuration for a browser service.
func (u *Upstream) Register(
	ctx context.Context,
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

	browserConfig := serviceConfig.GetBrowserService()
	if browserConfig == nil {
		return "", nil, nil, fmt.Errorf("browser service config is nil")
	}

	// Initialize Browser Service
	// Defaults
	width := int(browserConfig.GetViewportWidth())
	if width == 0 {
		width = 1920
	}
	height := int(browserConfig.GetViewportHeight())
	if height == 0 {
		height = 1080
	}
	headless := browserConfig.GetHeadless()
	// Default is true, but bool defaults to false in proto3.
	// To strictly follow "default true", we might need logic.
	// But in proto, we used `bool headless = 2;`.
	// If the user wants headless, they should set it.
	// Wait, the comment said "Default is true".
	// If the user omits it, it's false.
	// Let's assume explicit config for now, or assume "true" if unset?
	// Can't distinguish unset from false easily without optional.
	// Let's stick to false (headful) being default if unspecified, OR
	// Force it to true if unset?
	// Most automation wants headless.
	// Let's assume if the user didn't specify, we want headless.
	// But `headless` field is bool.
	// Let's just use the value provided. If they want headless, they set true.

	endpoint := browserConfig.GetEndpoint()

	svc, err := NewService(ctx, endpoint, headless, browserConfig.GetUserAgent(), width, height)
	if err != nil {
		return "", nil, nil, fmt.Errorf("failed to create browser service: %w", err)
	}

	u.mu.Lock()
	u.services = append(u.services, svc)
	u.mu.Unlock()

	timeout := 30 * time.Second
	if browserConfig.GetTimeout() != nil {
		timeout = browserConfig.GetTimeout().AsDuration()
	}

	tools, err := createTools(serviceID, svc, timeout, toolManager, serviceConfig)
	if err != nil {
		return "", nil, nil, err
	}

	log.Info("Registered browser service", "serviceID", serviceID, "tools", len(tools))
	return serviceID, tools, nil, nil
}
