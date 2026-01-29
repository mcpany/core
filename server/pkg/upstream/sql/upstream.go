// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package sql

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	// Import drivers for SQL upstream.
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "modernc.org/sqlite"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/prompt"
	"github.com/mcpany/core/server/pkg/resource"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/mcpany/core/server/pkg/util"
)

// Upstream implements the upstream.Upstream interface for SQL databases.
type Upstream struct {
	db *sql.DB
	mu sync.Mutex
}

// NewUpstream creates a new SQL upstream.
//
// Returns the result.
func NewUpstream() *Upstream {
	return &Upstream{}
}

// Shutdown closes the database connection.
//
// _ is an unused parameter.
//
// Returns an error if the operation fails.
func (u *Upstream) Shutdown(_ context.Context) error {
	u.mu.Lock()
	defer u.mu.Unlock()
	if u.db != nil {
		return u.db.Close()
	}
	return nil
}

func ptr(s string) *string {
	return &s
}

// Register discovers and registers tools from the SQL configuration.
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
	u.mu.Lock()
	defer u.mu.Unlock()

	sqlConfig := serviceConfig.GetSqlService()
	if sqlConfig == nil {
		return "", nil, nil, fmt.Errorf("sql service config is nil")
	}

	if u.db != nil {
		_ = u.db.Close()
	}

	var err error
	u.db, err = sql.Open(sqlConfig.GetDriver(), sqlConfig.GetDsn())
	if err != nil {
		return "", nil, nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := u.db.PingContext(ctx); err != nil {
		_ = u.db.Close()
		return "", nil, nil, fmt.Errorf("failed to ping database: %w", err)
	}

	toolDefs := make([]*configv1.ToolDefinition, 0, len(sqlConfig.GetCalls()))

	for id, callDef := range sqlConfig.GetCalls() {
		toolName := id
		sanitizedToolName, err := util.SanitizeToolName(toolName)
		if err != nil {
			return "", nil, nil, fmt.Errorf("invalid tool name %s: %w", toolName, err)
		}

		t := v1.Tool_builder{
			Name:         ptr(sanitizedToolName),
			Description:  ptr(fmt.Sprintf("Execute SQL query: %s", id)),
			InputSchema:  callDef.GetInputSchema(),
			OutputSchema: callDef.GetOutputSchema(),
			ServiceId:    ptr(serviceConfig.GetId()),
			Tags:         []string{"upstream:sql"},
		}.Build()

		sqlTool := NewTool(t, u.db, callDef, serviceConfig.GetCallPolicies(), id)

		if err := toolManager.AddTool(sqlTool); err != nil {
			return "", nil, nil, fmt.Errorf("failed to add tool %s: %w", toolName, err)
		}

		toolDefs = append(toolDefs, configv1.ToolDefinition_builder{
			Name:        ptr(sanitizedToolName),
			Description: ptr(t.GetDescription()),
			ServiceId:   ptr(serviceConfig.GetId()),
			InputSchema: callDef.GetInputSchema(),
			CallId:      ptr(id),
		}.Build())
	}

	return serviceConfig.GetId(), toolDefs, nil, nil
}
