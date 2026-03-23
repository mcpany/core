import re

with open('server/pkg/tool/types.go', 'r') as f:
    content = f.read()

content += '''
// bufferPool is a sync.Pool for bytes.Buffer to reduce allocations.
//
// Summary: Provides a thread-safe pool of bytes.Buffer objects to minimize heap allocations during tool execution.
// Parameters: None
// Returns: A sync.Pool instance.
// Errors: None
// Side Effects: Modifies memory allocations.
var bufferPool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

// getBuffer returns a bytes.Buffer from the pool.
//
// Summary: Retrieves a bytes.Buffer from the shared bufferPool.
// Parameters: None
// Returns:
//   - *bytes.Buffer: A pointer to a ready-to-use bytes.Buffer.
// Errors: None
// Side Effects: None
func getBuffer() *bytes.Buffer {
	return bufferPool.Get().(*bytes.Buffer)
}

// putBuffer resets and returns a bytes.Buffer to the pool.
//
// Summary: Resets the provided bytes.Buffer and returns it to the bufferPool for reuse.
// Parameters:
//   - buf (*bytes.Buffer): The buffer to return to the pool.
// Returns: None
// Errors: None
// Side Effects:
//   - Clears the contents of the provided buffer.
func putBuffer(buf *bytes.Buffer) {
	buf.Reset()
	bufferPool.Put(buf)
}
'''

with open('server/pkg/tool/types.go', 'w') as f:
    f.write(content)
