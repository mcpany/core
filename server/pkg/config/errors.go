// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import "fmt"

// ActionableError is an error that includes a suggestion for fixing the issue.
type ActionableError struct {
	Err        error
	Suggestion string
}

func (e *ActionableError) Error() string {
	return fmt.Sprintf("%v\n\t-> Fix: %s", e.Err, e.Suggestion)
}

func (e *ActionableError) Unwrap() error {
	return e.Err
}
