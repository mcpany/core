// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package filesystem

import (
	"context"
)

<<<<<<< HEAD
// filesystemToolDef represents a filesystem tool definition.
//
// Summary: Filesystem tool definition.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
=======
>>>>>>> 4f039895e (⚡ Bolt: Render Optimization for System Status Banner (#6544))
type filesystemToolDef struct {
	Name        string
	Description string
	Input       map[string]interface{}
	Output      map[string]interface{}
	Handler     func(ctx context.Context, args map[string]interface{}) (map[string]interface{}, error)
}
