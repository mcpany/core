// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// BlackboardTool acts as a shared Key-Value store tool.
//
// Summary: Blackboard Tool for agents.
type BlackboardTool struct {
	baseTool
	db *sql.DB
}

// NewBlackboardTool creates a new BlackboardTool.
//
// Summary: Initializes a new BlackboardTool.
//
// Parameters:
//   - dbPath: string. The path to the SQLite database.
//
// Returns:
//   - *BlackboardTool: The initialized tool.
//   - error: Error if initialization fails.
func NewBlackboardTool(dbPath string) (*BlackboardTool, error) {
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory for blackboard db: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open blackboard db: %w", err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS blackboard (
		key TEXT PRIMARY KEY,
		value TEXT NOT NULL
	)`)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create blackboard table: %w", err)
	}

	return &BlackboardTool{
		db: db,
	}, nil
}

// Close closes the underlying database.
func (b *BlackboardTool) Close() error {
	return b.db.Close()
}

// Get reads a value from the blackboard.
func (b *BlackboardTool) Get(ctx context.Context, key string) (string, error) {
	var value string
	err := b.db.QueryRowContext(ctx, "SELECT value FROM blackboard WHERE key = ?", key).Scan(&value)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", fmt.Errorf("failed to read from blackboard: %w", err)
	}
	return value, nil
}

// Set writes a value to the blackboard.
func (b *BlackboardTool) Set(ctx context.Context, key, value string) error {
	_, err := b.db.ExecContext(ctx, "INSERT OR REPLACE INTO blackboard (key, value) VALUES (?, ?)", key, value)
	if err != nil {
		return fmt.Errorf("failed to write to blackboard: %w", err)
	}
	return nil
}

// Execute runs the tool.
func (b *BlackboardTool) Execute(ctx context.Context, req *ExecutionRequest) (any, error) {
	operation, ok := req.Arguments["operation"].(string)
	if !ok {
		return nil, fmt.Errorf("missing operation argument")
	}

	key, ok := req.Arguments["key"].(string)
	if !ok {
		return nil, fmt.Errorf("missing key argument")
	}

	switch operation {
	case "get":
		val, err := b.Get(ctx, key)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"value": val}, nil
	case "set":
		value, ok := req.Arguments["value"].(string)
		if !ok {
			return nil, fmt.Errorf("missing value argument for set operation")
		}
		err := b.Set(ctx, key, value)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"status": "success"}, nil
	default:
		return nil, fmt.Errorf("unknown operation: %s", operation)
	}
}

// Definition returns the tool definition.
func (b *BlackboardTool) Definition() *mcp.Tool {
	return &mcp.Tool{
		Name:        "blackboard",
		Description: "A shared Key-Value store (Blackboard) for agent coordination and state sharing. Use 'get' to retrieve a value by key, and 'set' to store a value by key.",
		InputSchema: struct {
			Type       string                   `json:"type"`
			Properties map[string]interface{}   `json:"properties,omitempty"`
			Required   []string                 `json:"required,omitempty"`
		}{
			Type: "object",
			Properties: map[string]interface{}{
				"operation": map[string]interface{}{
					"type":        "string",
					"description": "The operation to perform: 'get' or 'set'",
					"enum":        []string{"get", "set"},
				},
				"key": map[string]interface{}{
					"type":        "string",
					"description": "The key to read or write",
				},
				"value": map[string]interface{}{
					"type":        "string",
					"description": "The value to write (only required for 'set')",
				},
			},
			Required: []string{"operation", "key"},
		},
	}
}
