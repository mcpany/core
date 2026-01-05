package provider

import (
	"io"

	"github.com/spf13/afero"
)

// Provider defines the interface for a filesystem provider.
type Provider interface {
	io.Closer

	// GetFs returns the underlying afero.Fs.
	GetFs() afero.Fs

	// ResolvePath resolves a virtual path to the actual path expected by the filesystem.
	ResolvePath(virtualPath string) (string, error)
}
