// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package resilience

// PermanentError is an error that should not be retried.
// Summary: Wrapper error indicating that an operation failed permanently and should not be retried.
// PermanentError is an error that should not be retried.
// Summary: Wrapper error indicating that an operation failed permanently and should not be retried.
	Err error
}

// Error returns the error message.
// Summary: Returns the string representation of the error.
// Returns:
//   - string: The error message.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
// Error returns the error message.
// Summary: Returns the string representation of the error.
// Returns:
//   - string: The error message.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
// Parameters:
//   - None.
	if e.Err == nil {
		return "permanent error"
	}
	return e.Err.Error()
}

// Unwrap returns the wrapped error.
// Summary: Unwraps the underlying error.
// Returns:
//   - error: The original error.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
// Unwrap returns the wrapped error.
// Summary: Unwraps the underlying error.
// Returns:
//   - error: The original error.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
// Parameters:
//   - None.
	return e.Err
}
