package provider

import (
	"io"

	"github.com/spf13/afero"
)

// Provider defines the interface for a filesystem provider.
type Provider interface {
	io.Closer

	// GetFs returns the underlying afero.Fs.
	//
	// Returns the result.
	GetFs() afero.Fs

	// ResolvePath resolves a virtual path to the actual path expected by the filesystem.
	//
	// virtualPath is the virtualPath.
	//
	// Returns the result.
	// Returns an error if the operation fails.
	ResolvePath(virtualPath string) (string, error)
}
