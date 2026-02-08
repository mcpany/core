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
// Summary: provides access to files in a Google Cloud Storage bucket.
type GcsProvider struct {
	fs     afero.Fs
	client *storage.Client
}

var newStorageClient = storage.NewClient

// NewGcsProvider creates a new GcsProvider from the given configuration.
//
// Summary: creates a new GcsProvider from the given configuration.
//
// Parameters:
//   - _: context.Context. The _.
//   - config: *configv1.GcsFs. The config.
//
// Returns:
//   - *GcsProvider: The *GcsProvider.
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
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

// GetFs returns the underlying filesystem.
//
// Summary: returns the underlying filesystem.
//
// Parameters:
//   None.
//
// Returns:
//   - afero.Fs: The afero.Fs.
func (p *GcsProvider) GetFs() afero.Fs {
	return p.fs
}

// ResolvePath resolves the virtual path to a real path in the bucket.
//
// Summary: resolves the virtual path to a real path in the bucket.
//
// Parameters:
//   - virtualPath: string. The virtualPath.
//
// Returns:
//   - string: The string.
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
func (p *GcsProvider) ResolvePath(virtualPath string) (string, error) {
	// Same as S3
	cleanPath := path.Clean("/" + virtualPath)
	cleanPath = strings.TrimPrefix(cleanPath, "/")

	if cleanPath == "" || cleanPath == "." {
		return "", fmt.Errorf("invalid path")
	}
	return cleanPath, nil
}

// Close closes the GCS client.
//
// Summary: closes the GCS client.
//
// Parameters:
//   None.
//
// Returns:
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
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

// Create creates a file in the filesystem, returning the file and an error, if any happens.
//
// Summary: creates a file in the filesystem, returning the file and an error, if any happens.
//
// Parameters:
//   - name: string. The name.
//
// Returns:
//   - afero.File: The afero.File.
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
func (fs *gcsFs) Create(name string) (afero.File, error) {
	return fs.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
}

// Mkdir creates a directory in the filesystem, returning an error, if any happens.
//
// Summary: creates a directory in the filesystem, returning an error, if any happens.
//
// Parameters:
//   - _: string. The _.
//   - _: os.FileMode. The _.
//
// Returns:
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
func (fs *gcsFs) Mkdir(_ string, _ os.FileMode) error {
	return nil // Flat namespace
}

// MkdirAll creates a directory path and all parents that does not exist for a given name.
//
// Summary: creates a directory path and all parents that does not exist for a given name.
//
// Parameters:
//   - _: string. The _.
//   - _: os.FileMode. The _.
//
// Returns:
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
func (fs *gcsFs) MkdirAll(_ string, _ os.FileMode) error {
	return nil // Flat namespace
}

// Open opens a file, returning it or an error, if any happens.
//
// Summary: opens a file, returning it or an error, if any happens.
//
// Parameters:
//   - name: string. The name.
//
// Returns:
//   - afero.File: The afero.File.
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
func (fs *gcsFs) Open(name string) (afero.File, error) {
	return fs.OpenFile(name, os.O_RDONLY, 0)
}

// OpenFile opens a file using the given flags and the given mode.
//
// Summary: opens a file using the given flags and the given mode.
//
// Parameters:
//   - name: string. The name.
//   - flag: int. The flag.
//   - _: os.FileMode. The _.
//
// Returns:
//   - afero.File: The afero.File.
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
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

// Remove removes a file identified by name, returning an error, if any happens.
//
// Summary: removes a file identified by name, returning an error, if any happens.
//
// Parameters:
//   - name: string. The name.
//
// Returns:
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
func (fs *gcsFs) Remove(name string) error {
	return fs.client.Bucket(fs.bucket).Object(name).Delete(fs.ctx)
}

// RemoveAll removes a directory path and any children it contains.
//
// Summary: removes a directory path and any children it contains.
//
// Parameters:
//   - path: string. The path.
//
// Returns:
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
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

// Rename renames a file.
//
// Summary: renames a file.
//
// Parameters:
//   - oldname: string. The oldname.
//   - newname: string. The newname.
//
// Returns:
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
func (fs *gcsFs) Rename(oldname, newname string) error {
	src := fs.client.Bucket(fs.bucket).Object(oldname)
	dst := fs.client.Bucket(fs.bucket).Object(newname)

	if _, err := dst.CopierFrom(src).Run(fs.ctx); err != nil {
		return err
	}
	return src.Delete(fs.ctx)
}

// Stat returns a FileInfo describing the named file, or an error, if any happens.
//
// Summary: returns a FileInfo describing the named file, or an error, if any happens.
//
// Parameters:
//   - name: string. The name.
//
// Returns:
//   - os.FileInfo: The os.FileInfo.
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
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

// Name returns the name of this file system.
//
// Summary: returns the name of this file system.
//
// Parameters:
//   None.
//
// Returns:
//   - string: The string.
func (fs *gcsFs) Name() string {
	return "gcs"
}

// Chmod changes the mode of the named file to mode.
//
// Summary: changes the mode of the named file to mode.
//
// Parameters:
//   - _: string. The _.
//   - _: os.FileMode. The _.
//
// Returns:
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
func (fs *gcsFs) Chmod(_ string, _ os.FileMode) error {
	return nil // Not supported
}

// Chown changes the uid and gid of the named file.
//
// Summary: changes the uid and gid of the named file.
//
// Parameters:
//   - _: string. The _.
//   - _: int. The _.
//   - _: int. The _.
//
// Returns:
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
func (fs *gcsFs) Chown(_ string, _, _ int) error {
	return nil // Not supported
}

// Chtimes changes the access and modification times of the named file.
//
// Summary: changes the access and modification times of the named file.
//
// Parameters:
//   - _: string. The _.
//   - _: time.Time. The _.
//   - _: time.Time. The _.
//
// Returns:
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
func (fs *gcsFs) Chtimes(_ string, _, _ time.Time) error {
	return nil // Not supported
}

type gcsFile struct {
	fs     *gcsFs
	name   string
	reader *storage.Reader
	writer *storage.Writer
}

// Close closes the file.
//
// Summary: closes the file.
//
// Parameters:
//   None.
//
// Returns:
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
func (f *gcsFile) Close() error {
	if f.writer != nil {
		return f.writer.Close()
	}
	if f.reader != nil {
		return f.reader.Close()
	}
	return nil
}

// Read reads up to len(b) bytes from the File.
//
// Summary: reads up to len(b) bytes from the File.
//
// Parameters:
//   - p: []byte. The p.
//
// Returns:
//   - n: int. The int.
//   - err: error. An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
func (f *gcsFile) Read(p []byte) (n int, err error) {
	if f.reader == nil {
		return 0, fmt.Errorf("file not opened for reading")
	}
	return f.reader.Read(p)
}

// ReadAt reads len(b) bytes from the File starting at byte offset off.
//
// Summary: reads len(b) bytes from the File starting at byte offset off.
//
// Parameters:
//   - p: []byte. The p.
//   - off: int64. The off.
//
// Returns:
//   - n: int. The int.
//   - err: error. An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
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

// Seek sets the offset for the next Read or Write to offset, interpreted according to whence.
//
// Summary: sets the offset for the next Read or Write to offset, interpreted according to whence.
//
// Parameters:
//   - _: int64. The _.
//   - _: int. The _.
//
// Returns:
//   - int64: The int64.
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
func (f *gcsFile) Seek(_ int64, _ int) (int64, error) {
	return 0, fmt.Errorf("seek not supported")
}

// Write writes len(b) bytes to the File.
//
// Summary: writes len(b) bytes to the File.
//
// Parameters:
//   - p: []byte. The p.
//
// Returns:
//   - n: int. The int.
//   - err: error. An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
func (f *gcsFile) Write(p []byte) (n int, err error) {
	if f.writer == nil {
		return 0, fmt.Errorf("file not opened for writing")
	}
	return f.writer.Write(p)
}

// WriteAt writes len(b) bytes to the File starting at byte offset off.
//
// Summary: writes len(b) bytes to the File starting at byte offset off.
//
// Parameters:
//   - _: []byte. The _.
//   - _: int64. The _.
//
// Returns:
//   - n: int. The int.
//   - err: error. An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
func (f *gcsFile) WriteAt(_ []byte, _ int64) (n int, err error) {
	return 0, fmt.Errorf("writeat not supported")
}

// Name returns the name of the file as presented to Open.
//
// Summary: returns the name of the file as presented to Open.
//
// Parameters:
//   None.
//
// Returns:
//   - string: The string.
func (f *gcsFile) Name() string {
	return f.name
}

// Readdir reads the contents of the directory associated with file and returns.
//
// Summary: reads the contents of the directory associated with file and returns.
//
// Parameters:
//   - _: int. The _.
//
// Returns:
//   - []os.FileInfo: The []os.FileInfo.
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
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

// Readdirnames reads and returns a slice of names from the directory f.
//
// Summary: reads and returns a slice of names from the directory f.
//
// Parameters:
//   - n: int. The n.
//
// Returns:
//   - []string: The []string.
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
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

// Stat returns the FileInfo structure describing file.
//
// Summary: returns the FileInfo structure describing file.
//
// Parameters:
//   None.
//
// Returns:
//   - os.FileInfo: The os.FileInfo.
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
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

// Sync commits the current contents of the file to stable storage.
//
// Summary: commits the current contents of the file to stable storage.
//
// Parameters:
//   None.
//
// Returns:
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
func (f *gcsFile) Sync() error {
	return nil
}

// Truncate changes the size of the file.
//
// Summary: changes the size of the file.
//
// Parameters:
//   - _: int64. The _.
//
// Returns:
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
func (f *gcsFile) Truncate(_ int64) error {
	return fmt.Errorf("truncate not supported")
}

// WriteString is like Write, but writes the contents of string s rather than a slice of bytes.
//
// Summary: is like Write, but writes the contents of string s rather than a slice of bytes.
//
// Parameters:
//   - s: string. The s.
//
// Returns:
//   - ret: int. The int.
//   - err: error. An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
func (f *gcsFile) WriteString(s string) (ret int, err error) {
	return f.Write([]byte(s))
}

type gcsFileInfo struct {
	name    string
	size    int64
	modTime time.Time
	isDir   bool
}

// Name returns the base name of the file.
//
// Summary: returns the base name of the file.
//
// Parameters:
//   None.
//
// Returns:
//   - string: The string.
func (fi *gcsFileInfo) Name() string {
	return fi.name
}

// Size returns the length in bytes for regular files; system-dependent for others.
//
// Summary: returns the length in bytes for regular files; system-dependent for others.
//
// Parameters:
//   None.
//
// Returns:
//   - int64: The int64.
func (fi *gcsFileInfo) Size() int64 {
	return fi.size
}

// Mode returns file mode bits.
//
// Summary: returns file mode bits.
//
// Parameters:
//   None.
//
// Returns:
//   - os.FileMode: The os.FileMode.
func (fi *gcsFileInfo) Mode() os.FileMode {
	if fi.isDir {
		return os.ModeDir | 0755
	}
	return 0644
}

// ModTime returns the modification time.
//
// Summary: returns the modification time.
//
// Parameters:
//   None.
//
// Returns:
//   - time.Time: The time.Time.
func (fi *gcsFileInfo) ModTime() time.Time {
	return fi.modTime
}

// IsDir returns true if the file is a directory.
//
// Summary: returns true if the file is a directory.
//
// Parameters:
//   None.
//
// Returns:
//   - bool: The bool.
func (fi *gcsFileInfo) IsDir() bool {
	return fi.isDir
}

// Sys returns underlying data source (can return nil).
//
// Summary: returns underlying data source (can return nil).
//
// Parameters:
//   None.
//
// Returns:
//   - interface{}: The interface{}.
func (fi *gcsFileInfo) Sys() interface{} {
	return nil
}
