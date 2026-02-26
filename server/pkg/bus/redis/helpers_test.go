package redis

import (
	"bytes"
	"sync"
)

// ThreadSafeBuffer is a bytes.Buffer that is safe for concurrent use.
// Copied from server_test.go for package-local usage in tests
type ThreadSafeBuffer struct {
	b bytes.Buffer
	m sync.Mutex
}

func (b *ThreadSafeBuffer) Write(p []byte) (n int, err error) {
	b.m.Lock()
	defer b.m.Unlock()
	return b.b.Write(p)
}

func (b *ThreadSafeBuffer) String() string {
	b.m.Lock()
	defer b.m.Unlock()
	return b.b.String()
}
