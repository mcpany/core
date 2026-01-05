package provider

import (
	"path/filepath"

	"github.com/spf13/afero"
)

type TmpfsProvider struct {
	fs afero.Fs
}

func NewTmpfsProvider() *TmpfsProvider {
	return &TmpfsProvider{
		fs: afero.NewMemMapFs(),
	}
}

func (p *TmpfsProvider) GetFs() afero.Fs {
	return p.fs
}

func (p *TmpfsProvider) ResolvePath(virtualPath string) (string, error) {
	// For MemMapFs, just clean the path. It's virtual.
	return filepath.Clean(virtualPath), nil
}

func (p *TmpfsProvider) Close() error {
	return nil
}
