package provider

import (
	"archive/zip"
	"fmt"
	"os"
	"path/filepath"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/spf13/afero"
	"github.com/spf13/afero/zipfs"
)

type ZipProvider struct {
	fs     afero.Fs
	closer *os.File
}

func NewZipProvider(config *configv1.ZipFs) (*ZipProvider, error) {
	f, err := os.Open(config.GetFilePath())
	if err != nil {
		return nil, fmt.Errorf("failed to open zip file: %w", err)
	}

	fi, err := f.Stat()
	if err != nil {
		_ = f.Close()
		return nil, fmt.Errorf("failed to stat zip file: %w", err)
	}

	zr, err := zip.NewReader(f, fi.Size())
	if err != nil {
		_ = f.Close()
		return nil, fmt.Errorf("failed to create zip reader: %w", err)
	}

	fs := zipfs.New(zr)

	return &ZipProvider{
		fs:     fs,
		closer: f,
	}, nil
}

func (p *ZipProvider) GetFs() afero.Fs {
	return p.fs
}

func (p *ZipProvider) ResolvePath(virtualPath string) (string, error) {
	// For ZipFs, just clean the path. It's virtual (based on zip contents).
	return filepath.Clean(virtualPath), nil
}

func (p *ZipProvider) Close() error {
	if p.closer != nil {
		return p.closer.Close()
	}
	return nil
}
