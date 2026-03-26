// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import "context"

// SeedDataPublic exposes seedData for tests.
func (a *Application) SeedDataPublic(ctx context.Context, req SeedRequest) error {
	return a.seedData(ctx, req)
}
