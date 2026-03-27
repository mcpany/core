// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package provider implements filesystem providers.
package provider

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/spf13/afero"
	"google.golang.org/api/iterator"
)

// GcsProvider provides access to files in a Google Cloud Storage bucket.
//
// Summary: Represents a GcsProvider.
type GcsProvider struct {
	fs     afero.Fs
	client *storage.Client
}

var newStorageClient = storage.NewClient

// NewGcsProvider provides newgcsprovider functionality.
//
// Summary: NewGcsProvider.
//
// Parameters.
//   - _: The parameter.
//   - config: The parameter.
//   - error: The parameter.
//
// Returns.
//   - None.
func NewGcsProvider(_ context.Context, config *configv1.GcsFs) (*GcsProvider, error) {
	if config == nil {
		return nil, fmt.Errorf("gcs config is nil")
	}

	client, err := newStorageClient(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to create gcs client: %w", err)
	}

	return &GcsProvider{
		fs:     &gcsFs{client: client, bucket: config.GetBucket(), ctx: context.Background()},
		client: client,
	}, nil
}

// GetFs provides getfs functionality.
//
// Summary: GetFs.
//
// Parameters.
//   - None.
//
// Returns.
//   - result: The result.
func (p *GcsProvider) GetFs() afero.Fs {
	return p.fs
}

// ResolvePath provides resolvepath functionality.
//
// Summary: ResolvePath.
//
// Parameters.
//   - virtualPath: The parameter.
//   - error: The parameter.
//
// Returns.
//   - None.
func (p *GcsProvider) ResolvePath(virtualPath string) (string, error) {
	// Same as S3
	cleanPath := path.Clean("/" + virtualPath)
	cleanPath = strings.TrimPrefix(cleanPath, "/")

	if cleanPath == "" || cleanPath == "." {
		return "", fmt.Errorf("invalid path")
	}
	return cleanPath, nil
}

// Close provides close functionality.
//
// Summary: Close.
//
// Parameters.
//   - None.
//
// Returns.
//   - result: The result.
func (p *GcsProvider) Close() error {
	if p.client != nil {
		return p.client.Close()
	}
	return nil
}

// gcsFs implementation copy from original gcs.go

type gcsFs struct {
	client *storage.Client
	bucket string
	ctx    context.Context
}

// Create provides create functionality.
//
// Summary: Create.
//
// Parameters.
//   - name: The parameter.
//   - error: The parameter.
//
// Returns.
//   - None.
func (fs *gcsFs) Create(name string) (afero.File, error) {
	return fs.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
}

// Mkdir provides mkdir functionality.
//
// Summary: Mkdir.
//
// Parameters.
//   - _: The parameter.
//   - _: The parameter.
//
// Returns.
//   - result: The result.
func (fs *gcsFs) Mkdir(_ string, _ os.FileMode) error {
	return nil // Flat namespace
}

// MkdirAll provides mkdirall functionality.
//
// Summary: MkdirAll.
//
// Parameters.
//   - _: The parameter.
//   - _: The parameter.
//
// Returns.
//   - result: The result.
func (fs *gcsFs) MkdirAll(_ string, _ os.FileMode) error {
	return nil // Flat namespace
}

// Open provides open functionality.
//
// Summary: Open.
//
// Parameters.
//   - name: The parameter.
//   - error: The parameter.
//
// Returns.
//   - None.
func (fs *gcsFs) Open(name string) (afero.File, error) {
	return fs.OpenFile(name, os.O_RDONLY, 0)
}

// OpenFile provides openfile functionality.
//
// Summary: OpenFile.
//
// Parameters.
//   - name: The parameter.
//   - flag: The parameter.
//   - _: The parameter.
//   - error: The parameter.
//
// Returns.
//   - None.
func (fs *gcsFs) OpenFile(name string, flag int, _ os.FileMode) (afero.File, error) {
	f := &gcsFile{
		fs:   fs,
		name: name,
	}

	if flag&os.O_RDWR != 0 || flag&os.O_WRONLY != 0 {
		// Write mode
		wc := fs.client.Bucket(fs.bucket).Object(name).NewWriter(fs.ctx)
		f.writer = wc
		return f, nil
	}

	// Read mode
	rc, err := fs.client.Bucket(fs.bucket).Object(name).NewReader(fs.ctx)
	if err != nil {
		if errors.Is(err, storage.ErrObjectNotExist) {
			return nil, os.ErrNotExist
		}
		return nil, err
	}
	f.reader = rc
	return f, nil
}

// Remove provides remove functionality.
//
// Summary: Remove.
//
// Parameters.
//   - name: The parameter.
//
// Returns.
//   - result: The result.
func (fs *gcsFs) Remove(name string) error {
	return fs.client.Bucket(fs.bucket).Object(name).Delete(fs.ctx)
}

// RemoveAll provides removeall functionality.
//
// Summary: RemoveAll.
//
// Parameters.
//   - path: The parameter.
//
// Returns.
//   - result: The result.
func (fs *gcsFs) RemoveAll(path string) error {
	// Delete everything with prefix
	it := fs.client.Bucket(fs.bucket).Objects(fs.ctx, &storage.Query{Prefix: path})
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return err
		}
		if err := fs.client.Bucket(fs.bucket).Object(attrs.Name).Delete(fs.ctx); err != nil {
			return err
		}
	}
	return nil
}

// Rename provides rename functionality.
//
// Summary: Rename.
//
// Parameters.
//   - oldname: The parameter.
//   - newname: The parameter.
//
// Returns.
//   - result: The result.
func (fs *gcsFs) Rename(oldname, newname string) error {
	src := fs.client.Bucket(fs.bucket).Object(oldname)
	dst := fs.client.Bucket(fs.bucket).Object(newname)

	if _, err := dst.CopierFrom(src).Run(fs.ctx); err != nil {
		return err
	}
	return src.Delete(fs.ctx)
}

// Stat provides stat functionality.
//
// Summary: Stat.
//
// Parameters.
//   - name: The parameter.
//   - error: The parameter.
//
// Returns.
//   - None.
func (fs *gcsFs) Stat(name string) (os.FileInfo, error) {
	attrs, err := fs.client.Bucket(fs.bucket).Object(name).Attrs(fs.ctx)
	if err != nil {
		if errors.Is(err, storage.ErrObjectNotExist) {
			return nil, os.ErrNotExist
		}
		return nil, err
	}
	return &gcsFileInfo{
		name:    path.Base(attrs.Name),
		size:    attrs.Size,
		modTime: attrs.Updated,
		isDir:   false,
	}, nil
}

// Name provides name functionality.
//
// Summary: Name.
//
// Parameters.
//   - None.
//
// Returns.
//   - result: The result.
func (fs *gcsFs) Name() string {
	return "gcs"
}

// Chmod provides chmod functionality.
//
// Summary: Chmod.
//
// Parameters.
//   - _: The parameter.
//   - _: The parameter.
//
// Returns.
//   - result: The result.
func (fs *gcsFs) Chmod(_ string, _ os.FileMode) error {
	return nil // Not supported
}

// Chown provides chown functionality.
//
// Summary: Chown.
//
// Parameters.
//   - _: The parameter.
//   - _: The parameter.
//   - _: The parameter.
//
// Returns.
//   - result: The result.
func (fs *gcsFs) Chown(_ string, _, _ int) error {
	return nil // Not supported
}

// Chtimes provides chtimes functionality.
//
// Summary: Chtimes.
//
// Parameters.
//   - _: The parameter.
//   - _: The parameter.
//   - _: The parameter.
//
// Returns.
//   - result: The result.
func (fs *gcsFs) Chtimes(_ string, _, _ time.Time) error {
	return nil // Not supported
}

type gcsFile struct {
	fs     *gcsFs
	name   string
	reader *storage.Reader
	writer *storage.Writer
}

// Close provides close functionality.
//
// Summary: Close.
//
// Parameters.
//   - None.
//
// Returns.
//   - result: The result.
func (f *gcsFile) Close() error {
	if f.writer != nil {
		return f.writer.Close()
	}
	if f.reader != nil {
		return f.reader.Close()
	}
	return nil
}

// Read provides read functionality.
//
// Summary: Read.
//
// Parameters.
//   - p: The parameter.
//   - err: The parameter.
//
// Returns.
//   - None.
func (f *gcsFile) Read(p []byte) (n int, err error) {
	if f.reader == nil {
		return 0, fmt.Errorf("file not opened for reading")
	}
	return f.reader.Read(p)
}

// ReadAt provides readat functionality.
//
// Summary: ReadAt.
//
// Parameters.
//   - p: The parameter.
//   - off: The parameter.
//   - err: The parameter.
//
// Returns.
//   - None.
func (f *gcsFile) ReadAt(p []byte, off int64) (n int, err error) {
	// storage.Reader doesn't support ReadAt directly unless created with range?
	// But afero.File requires ReadAt.
	// We can create a new reader with range.
	rc, err := f.fs.client.Bucket(f.fs.bucket).Object(f.name).NewRangeReader(f.fs.ctx, off, int64(len(p)))
	if err != nil {
		return 0, err
	}
	defer func() { _ = rc.Close() }()
	return io.ReadFull(rc, p)
}

// Seek provides seek functionality.
//
// Summary: Seek.
//
// Parameters.
//   - _: The parameter.
//   - _: The parameter.
//   - error: The parameter.
//
// Returns.
//   - None.
func (f *gcsFile) Seek(_ int64, _ int) (int64, error) {
	return 0, fmt.Errorf("seek not supported")
}

// Write provides write functionality.
//
// Summary: Write.
//
// Parameters.
//   - p: The parameter.
//   - err: The parameter.
//
// Returns.
//   - None.
func (f *gcsFile) Write(p []byte) (n int, err error) {
	if f.writer == nil {
		return 0, fmt.Errorf("file not opened for writing")
	}
	return f.writer.Write(p)
}

// WriteAt provides writeat functionality.
//
// Summary: WriteAt.
//
// Parameters.
//   - _: The parameter.
//   - _: The parameter.
//   - err: The parameter.
//
// Returns.
//   - None.
func (f *gcsFile) WriteAt(_ []byte, _ int64) (n int, err error) {
	return 0, fmt.Errorf("writeat not supported")
}

// Name provides name functionality.
//
// Summary: Name.
//
// Parameters.
//   - None.
//
// Returns.
//   - result: The result.
func (f *gcsFile) Name() string {
	return f.name
}

// Readdir provides readdir functionality.
//
// Summary: Readdir.
//
// Parameters.
//   - _: The parameter.
//   - error: The parameter.
//
// Returns.
//   - None.
func (f *gcsFile) Readdir(_ int) ([]os.FileInfo, error) {
	// List objects with prefix name/
	prefix := f.name
	if !strings.HasSuffix(prefix, "/") && prefix != "" {
		prefix += "/"
	}
	if prefix == "/" {
		prefix = "" // Root
	}

	it := f.fs.client.Bucket(f.fs.bucket).Objects(f.fs.ctx, &storage.Query{
		Prefix:    prefix,
		Delimiter: "/",
	})

	var infos []os.FileInfo
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		if attrs.Prefix != "" {
			// Directory
			infos = append(infos, &gcsFileInfo{
				name:  path.Base(strings.TrimSuffix(attrs.Prefix, "/")),
				size:  0,
				isDir: true,
			})
		} else {
			infos = append(infos, &gcsFileInfo{
				name:    path.Base(attrs.Name),
				size:    attrs.Size,
				modTime: attrs.Updated,
				isDir:   false,
			})
		}
	}
	return infos, nil
}

// Readdirnames provides readdirnames functionality.
//
// Summary: Readdirnames.
//
// Parameters.
//   - n: The parameter.
//   - error: The parameter.
//
// Returns.
//   - None.
func (f *gcsFile) Readdirnames(n int) ([]string, error) {
	infos, err := f.Readdir(n)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(infos))
	for _, info := range infos {
		names = append(names, info.Name())
	}
	return names, nil
}

// Stat provides stat functionality.
//
// Summary: Stat.
//
// Parameters.
//   - ): The parameter.
//   - error: The parameter.
//
// Returns.
//   - None.
func (f *gcsFile) Stat() (os.FileInfo, error) {
	if f.reader != nil {
		return &gcsFileInfo{
			name:    f.name, // ReaderObjectAttrs doesn't always have name?
			size:    f.reader.Attrs.Size,
			modTime: f.reader.Attrs.LastModified,
			isDir:   false,
		}, nil
	}
	if f.writer != nil {
		// Writer doesn't have attrs until closed?
		return &gcsFileInfo{
			name:  f.name,
			size:  0,
			isDir: false,
		}, nil
	}
	// Fallback to stat
	return f.fs.Stat(f.name)
}

// Sync provides sync functionality.
//
// Summary: Sync.
//
// Parameters.
//   - None.
//
// Returns.
//   - result: The result.
func (f *gcsFile) Sync() error {
	return nil
}

// Truncate provides truncate functionality.
//
// Summary: Truncate.
//
// Parameters.
//   - _: The parameter.
//
// Returns.
//   - result: The result.
func (f *gcsFile) Truncate(_ int64) error {
	return fmt.Errorf("truncate not supported")
}

// WriteString provides writestring functionality.
//
// Summary: WriteString.
//
// Parameters.
//   - s: The parameter.
//   - err: The parameter.
//
// Returns.
//   - None.
func (f *gcsFile) WriteString(s string) (ret int, err error) {
	return f.Write([]byte(s))
}

type gcsFileInfo struct {
	name    string
	size    int64
	modTime time.Time
	isDir   bool
}

// Name provides name functionality.
//
// Summary: Name.
//
// Parameters.
//   - None.
//
// Returns.
//   - result: The result.
func (fi *gcsFileInfo) Name() string {
	return fi.name
}

// Size provides size functionality.
//
// Summary: Size.
//
// Parameters.
//   - None.
//
// Returns.
//   - result: The result.
func (fi *gcsFileInfo) Size() int64 {
	return fi.size
}

// Mode provides mode functionality.
//
// Summary: Mode.
//
// Parameters.
//   - None.
//
// Returns.
//   - result: The result.
func (fi *gcsFileInfo) Mode() os.FileMode {
	if fi.isDir {
		return os.ModeDir | 0755
	}
	return 0644
}

// ModTime provides modtime functionality.
//
// Summary: ModTime.
//
// Parameters.
//   - None.
//
// Returns.
//   - result: The result.
func (fi *gcsFileInfo) ModTime() time.Time {
	return fi.modTime
}

// IsDir provides isdir functionality.
//
// Summary: IsDir.
//
// Parameters.
//   - None.
//
// Returns.
//   - result: The result.
func (fi *gcsFileInfo) IsDir() bool {
	return fi.isDir
}

// Sys provides sys functionality.
//
// Summary: Sys.
//
// Parameters.
//   - None.
//
// Returns.
//   - result: The result.
func (fi *gcsFileInfo) Sys() interface{} {
	return nil
}
