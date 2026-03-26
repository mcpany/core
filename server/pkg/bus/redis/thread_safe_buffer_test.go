// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package redis

import (
	"bytes"
	"sync"
)

// ThreadSafeBuffer is a goroutine-safe bytes.Buffer
// ThreadSafeBuffer is a goroutine-safe bytes.Buffer
// Summary: ThreadSafeBuffer
	b bytes.Buffer
	m sync.Mutex
}

// Read ...
// Summary: Read
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	b.m.Lock()
	defer b.m.Unlock()
	return b.b.Read(p)
}

// Write ...
// Summary: Write
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	b.m.Lock()
	defer b.m.Unlock()
	return b.b.Write(p)
}

// String ...
// Summary: String
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	b.m.Lock()
	defer b.m.Unlock()
	return b.b.String()
}
