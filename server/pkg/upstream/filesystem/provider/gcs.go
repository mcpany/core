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

// NewGcsProvider creates a new gcs provider.
//
// Summary: Creates a new gcs provider.
//
// Parameters:
//   - _ (context.Context): Unused parameter.
//   - config (*configv1.GcsFs): The config.
//
// Returns:
//   - *GcsProvider: The result.
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
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

// GetFs retrieves the fs.
//
// Summary: Retrieves the fs.
//
// Parameters:
//   None.
//
// Returns:
//   - afero.Fs: The result.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (p *GcsProvider) GetFs() afero.Fs {
	return p.fs
}

// ResolvePath resolvePath resolve path.
//
// Summary: ResolvePath resolve path.
//
// Parameters:
//   - virtualPath (string): The virtual path.
//
// Returns:
//   - string: The result.
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
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

// Close close close.
//
// Summary: Close close.
//
// Parameters:
//   None.
//
// Returns:
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - None.
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

// Create persists the .
//
// Summary: Persists the .
//
// Parameters:
//   - name (string): The name.
//
// Returns:
//   - afero.File: The result.
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - None.
func (fs *gcsFs) Create(name string) (afero.File, error) {
	return fs.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
}

// Mkdir mkdir mkdir.
//
// Summary: Mkdir mkdir.
//
// Parameters:
//   - _ (string): Unused parameter.
//   - _ (os.FileMode): Unused parameter.
//
// Returns:
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - None.
func (fs *gcsFs) Mkdir(_ string, _ os.FileMode) error {
	return nil // Flat namespace
}

// MkdirAll mkdirAll mkdir all.
//
// Summary: MkdirAll mkdir all.
//
// Parameters:
//   - _ (string): Unused parameter.
//   - _ (os.FileMode): Unused parameter.
//
// Returns:
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - None.
func (fs *gcsFs) MkdirAll(_ string, _ os.FileMode) error {
	return nil // Flat namespace
}

// Open open open.
//
// Summary: Open open.
//
// Parameters:
//   - name (string): The name.
//
// Returns:
//   - afero.File: The result.
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - None.
func (fs *gcsFs) Open(name string) (afero.File, error) {
	return fs.OpenFile(name, os.O_RDONLY, 0)
}

// OpenFile openFile open file.
//
// Summary: OpenFile open file.
//
// Parameters:
//   - name (string): The name.
//   - flag (int): The flag.
//   - _ (os.FileMode): Unused parameter.
//
// Returns:
//   - afero.File: The result.
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
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

// Remove remove remove.
//
// Summary: Remove remove.
//
// Parameters:
//   - name (string): The name.
//
// Returns:
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - None.
func (fs *gcsFs) Remove(name string) error {
	return fs.client.Bucket(fs.bucket).Object(name).Delete(fs.ctx)
}

// RemoveAll removeAll remove all.
//
// Summary: RemoveAll remove all.
//
// Parameters:
//   - path (string): The path.
//
// Returns:
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - None.
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

// Rename rename rename.
//
// Summary: Rename rename.
//
// Parameters:
//   - oldname (unknown): The oldname.
//   - newname (string): The newname.
//
// Returns:
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - None.
func (fs *gcsFs) Rename(oldname, newname string) error {
	src := fs.client.Bucket(fs.bucket).Object(oldname)
	dst := fs.client.Bucket(fs.bucket).Object(newname)

	if _, err := dst.CopierFrom(src).Run(fs.ctx); err != nil {
		return err
	}
	return src.Delete(fs.ctx)
}

// Stat stat stat.
//
// Summary: Stat stat.
//
// Parameters:
//   - name (string): The name.
//
// Returns:
//   - os.FileInfo: The result.
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
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

// Name name name.
//
// Summary: Name name.
//
// Parameters:
//   None.
//
// Returns:
//   - string: The result.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (fs *gcsFs) Name() string {
	return "gcs"
}

// Chmod chmod chmod.
//
// Summary: Chmod chmod.
//
// Parameters:
//   - _ (string): Unused parameter.
//   - _ (os.FileMode): Unused parameter.
//
// Returns:
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - None.
func (fs *gcsFs) Chmod(_ string, _ os.FileMode) error {
	return nil // Not supported
}

// Chown chown chown.
//
// Summary: Chown chown.
//
// Parameters:
//   - _ (string): Unused parameter.
//   - _ (unknown): Unused parameter.
//   - _ (int): Unused parameter.
//
// Returns:
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - None.
func (fs *gcsFs) Chown(_ string, _, _ int) error {
	return nil // Not supported
}

// Chtimes chtimes chtimes.
//
// Summary: Chtimes chtimes.
//
// Parameters:
//   - _ (string): Unused parameter.
//   - _ (unknown): Unused parameter.
//   - _ (time.Time): Unused parameter.
//
// Returns:
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - None.
func (fs *gcsFs) Chtimes(_ string, _, _ time.Time) error {
	return nil // Not supported
}

type gcsFile struct {
	fs     *gcsFs
	name   string
	reader *storage.Reader
	writer *storage.Writer
}

// Close close close.
//
// Summary: Close close.
//
// Parameters:
//   None.
//
// Returns:
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - None.
func (f *gcsFile) Close() error {
	if f.writer != nil {
		return f.writer.Close()
	}
	if f.reader != nil {
		return f.reader.Close()
	}
	return nil
}

// Read read read.
//
// Summary: Read read.
//
// Parameters:
//   - p ([]byte): The p.
//
// Returns:
//   - int: The result.
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - None.
func (f *gcsFile) Read(p []byte) (n int, err error) {
	if f.reader == nil {
		return 0, fmt.Errorf("file not opened for reading")
	}
	return f.reader.Read(p)
}

// ReadAt readAt read at.
//
// Summary: ReadAt read at.
//
// Parameters:
//   - p ([]byte): The p.
//   - off (int64): The off.
//
// Returns:
//   - int: The result.
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
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

// Seek seek seek.
//
// Summary: Seek seek.
//
// Parameters:
//   - _ (int64): Unused parameter.
//   - _ (int): Unused parameter.
//
// Returns:
//   - int64: The result.
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - None.
func (f *gcsFile) Seek(_ int64, _ int) (int64, error) {
	return 0, fmt.Errorf("seek not supported")
}

// Write write write.
//
// Summary: Write write.
//
// Parameters:
//   - p ([]byte): The p.
//
// Returns:
//   - int: The result.
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - None.
func (f *gcsFile) Write(p []byte) (n int, err error) {
	if f.writer == nil {
		return 0, fmt.Errorf("file not opened for writing")
	}
	return f.writer.Write(p)
}

// WriteAt writeAt write at.
//
// Summary: WriteAt write at.
//
// Parameters:
//   - _ ([]byte): Unused parameter.
//   - _ (int64): Unused parameter.
//
// Returns:
//   - int: The result.
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - None.
func (f *gcsFile) WriteAt(_ []byte, _ int64) (n int, err error) {
	return 0, fmt.Errorf("writeat not supported")
}

// Name name name.
//
// Summary: Name name.
//
// Parameters:
//   None.
//
// Returns:
//   - string: The result.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (f *gcsFile) Name() string {
	return f.name
}

// Readdir readdir readdir.
//
// Summary: Readdir readdir.
//
// Parameters:
//   - _ (int): Unused parameter.
//
// Returns:
//   - []os.FileInfo: The result.
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
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

// Readdirnames readdirnames readdirnames.
//
// Summary: Readdirnames readdirnames.
//
// Parameters:
//   - n (int): The n.
//
// Returns:
//   - []string: The result.
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
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

// Stat stat stat.
//
// Summary: Stat stat.
//
// Parameters:
//   None.
//
// Returns:
//   - os.FileInfo: The result.
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
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

// Sync sync sync.
//
// Summary: Sync sync.
//
// Parameters:
//   None.
//
// Returns:
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - None.
func (f *gcsFile) Sync() error {
	return nil
}

// Truncate truncate truncate.
//
// Summary: Truncate truncate.
//
// Parameters:
//   - _ (int64): Unused parameter.
//
// Returns:
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - None.
func (f *gcsFile) Truncate(_ int64) error {
	return fmt.Errorf("truncate not supported")
}

// WriteString writeString write string.
//
// Summary: WriteString write string.
//
// Parameters:
//   - s (string): The s.
//
// Returns:
//   - int: The result.
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
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

// Name name name.
//
// Summary: Name name.
//
// Parameters:
//   None.
//
// Returns:
//   - string: The result.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (fi *gcsFileInfo) Name() string {
	return fi.name
}

// Size size size.
//
// Summary: Size size.
//
// Parameters:
//   None.
//
// Returns:
//   - int64: The result.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (fi *gcsFileInfo) Size() int64 {
	return fi.size
}

// Mode mode mode.
//
// Summary: Mode mode.
//
// Parameters:
//   None.
//
// Returns:
//   - os.FileMode: The result.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (fi *gcsFileInfo) Mode() os.FileMode {
	if fi.isDir {
		return os.ModeDir | 0755
	}
	return 0644
}

// ModTime modTime mod time.
//
// Summary: ModTime mod time.
//
// Parameters:
//   None.
//
// Returns:
//   - time.Time: The result.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (fi *gcsFileInfo) ModTime() time.Time {
	return fi.modTime
}

// IsDir isDir is dir.
//
// Summary: IsDir is dir.
//
// Parameters:
//   None.
//
// Returns:
//   - bool: The result.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (fi *gcsFileInfo) IsDir() bool {
	return fi.isDir
}

// Sys sys sys.
//
// Summary: Sys sys.
//
// Parameters:
//   None.
//
// Returns:
//   - interface: The result.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (fi *gcsFileInfo) Sys() interface{} {
	return nil
}
