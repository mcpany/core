// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package tool implements the tool subsystem.
package tool

import "errors"

// ErrToolNotFound is returned when a requested tool cannot be found.
var ErrToolNotFound = errors.New("unknown tool")
