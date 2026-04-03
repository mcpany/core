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

// GcsProvider represents the public GcsProvider entity.
//
// Summary: Defines the structured data model representing a provider.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
type GcsProvider struct {
	fs     afero.Fs
	client *storage.Client
}

var newStorageClient = storage.NewClient

// NewGcsProvider serves as a public interface for interacting with NewGcsProvider.
//
// Summary: Constructs and returns an initialized gcs provider ready for consumption.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
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

// GetFs serves as a public interface for interacting with GetFs.
//
// Summary: Fetches and returns the underlying fs from the system state.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (p *GcsProvider) GetFs() afero.Fs {
	return p.fs
}

// ResolvePath serves as a public interface for interacting with ResolvePath.
//
// Summary: Resolve the path appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (p *GcsProvider) ResolvePath(virtualPath string) (string, error) {
	// Same as S3
	cleanPath := path.Clean("/" + virtualPath)
	cleanPath = strings.TrimPrefix(cleanPath, "/")

	if cleanPath == "" || cleanPath == "." {
		return "", fmt.Errorf("invalid path")
	}
	return cleanPath, nil
}

// Close serves as a public interface for interacting with Close.
//
// Summary: Close the  appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
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

// Create serves as a public interface for interacting with Create.
//
// Summary: Create the  appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (fs *gcsFs) Create(name string) (afero.File, error) {
	return fs.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
}

// Mkdir serves as a public interface for interacting with Mkdir.
//
// Summary: Mkdir the  appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (fs *gcsFs) Mkdir(_ string, _ os.FileMode) error {
	return nil // Flat namespace
}

// MkdirAll serves as a public interface for interacting with MkdirAll.
//
// Summary: Mkdir the all appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (fs *gcsFs) MkdirAll(_ string, _ os.FileMode) error {
	return nil // Flat namespace
}

// Open serves as a public interface for interacting with Open.
//
// Summary: Open the  appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (fs *gcsFs) Open(name string) (afero.File, error) {
	return fs.OpenFile(name, os.O_RDONLY, 0)
}

// OpenFile serves as a public interface for interacting with OpenFile.
//
// Summary: Open the file appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
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

// Remove serves as a public interface for interacting with Remove.
//
// Summary: Remove the  appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (fs *gcsFs) Remove(name string) error {
	return fs.client.Bucket(fs.bucket).Object(name).Delete(fs.ctx)
}

// RemoveAll serves as a public interface for interacting with RemoveAll.
//
// Summary: Remove the all appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
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

// Rename serves as a public interface for interacting with Rename.
//
// Summary: Rename the  appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (fs *gcsFs) Rename(oldname, newname string) error {
	src := fs.client.Bucket(fs.bucket).Object(oldname)
	dst := fs.client.Bucket(fs.bucket).Object(newname)

	if _, err := dst.CopierFrom(src).Run(fs.ctx); err != nil {
		return err
	}
	return src.Delete(fs.ctx)
}

// Stat serves as a public interface for interacting with Stat.
//
// Summary: Stat the  appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
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

// Name serves as a public interface for interacting with Name.
//
// Summary: Name the  appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (fs *gcsFs) Name() string {
	return "gcs"
}

// Chmod serves as a public interface for interacting with Chmod.
//
// Summary: Chmod the  appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (fs *gcsFs) Chmod(_ string, _ os.FileMode) error {
	return nil // Not supported
}

// Chown serves as a public interface for interacting with Chown.
//
// Summary: Chown the  appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (fs *gcsFs) Chown(_ string, _, _ int) error {
	return nil // Not supported
}

// Chtimes serves as a public interface for interacting with Chtimes.
//
// Summary: Chtimes the  appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (fs *gcsFs) Chtimes(_ string, _, _ time.Time) error {
	return nil // Not supported
}

type gcsFile struct {
	fs     *gcsFs
	name   string
	reader *storage.Reader
	writer *storage.Writer
}

// Close serves as a public interface for interacting with Close.
//
// Summary: Close the  appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (f *gcsFile) Close() error {
	if f.writer != nil {
		return f.writer.Close()
	}
	if f.reader != nil {
		return f.reader.Close()
	}
	return nil
}

// Read serves as a public interface for interacting with Read.
//
// Summary: Read the  appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (f *gcsFile) Read(p []byte) (n int, err error) {
	if f.reader == nil {
		return 0, fmt.Errorf("file not opened for reading")
	}
	return f.reader.Read(p)
}

// ReadAt serves as a public interface for interacting with ReadAt.
//
// Summary: Read the at appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
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

// Seek serves as a public interface for interacting with Seek.
//
// Summary: Seek the  appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (f *gcsFile) Seek(_ int64, _ int) (int64, error) {
	return 0, fmt.Errorf("seek not supported")
}

// Write serves as a public interface for interacting with Write.
//
// Summary: Write the  appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (f *gcsFile) Write(p []byte) (n int, err error) {
	if f.writer == nil {
		return 0, fmt.Errorf("file not opened for writing")
	}
	return f.writer.Write(p)
}

// WriteAt serves as a public interface for interacting with WriteAt.
//
// Summary: Write the at appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (f *gcsFile) WriteAt(_ []byte, _ int64) (n int, err error) {
	return 0, fmt.Errorf("writeat not supported")
}

// Name serves as a public interface for interacting with Name.
//
// Summary: Name the  appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (f *gcsFile) Name() string {
	return f.name
}

// Readdir serves as a public interface for interacting with Readdir.
//
// Summary: Readdir the  appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
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

// Readdirnames serves as a public interface for interacting with Readdirnames.
//
// Summary: Readdirnames the  appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
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

// Stat serves as a public interface for interacting with Stat.
//
// Summary: Stat the  appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
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

// Sync serves as a public interface for interacting with Sync.
//
// Summary: Sync the  appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (f *gcsFile) Sync() error {
	return nil
}

// Truncate serves as a public interface for interacting with Truncate.
//
// Summary: Truncate the  appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (f *gcsFile) Truncate(_ int64) error {
	return fmt.Errorf("truncate not supported")
}

// WriteString serves as a public interface for interacting with WriteString.
//
// Summary: Write the string appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (f *gcsFile) WriteString(s string) (ret int, err error) {
	return f.Write([]byte(s))
}

type gcsFileInfo struct {
	name    string
	size    int64
	modTime time.Time
	isDir   bool
}

// Name serves as a public interface for interacting with Name.
//
// Summary: Name the  appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (fi *gcsFileInfo) Name() string {
	return fi.name
}

// Size serves as a public interface for interacting with Size.
//
// Summary: Size the  appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (fi *gcsFileInfo) Size() int64 {
	return fi.size
}

// Mode serves as a public interface for interacting with Mode.
//
// Summary: Mode the  appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (fi *gcsFileInfo) Mode() os.FileMode {
	if fi.isDir {
		return os.ModeDir | 0755
	}
	return 0644
}

// ModTime serves as a public interface for interacting with ModTime.
//
// Summary: Mod the time appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (fi *gcsFileInfo) ModTime() time.Time {
	return fi.modTime
}

// IsDir serves as a public interface for interacting with IsDir.
//
// Summary: Checks condition indicating whether the target is dir.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (fi *gcsFileInfo) IsDir() bool {
	return fi.isDir
}

// Sys serves as a public interface for interacting with Sys.
//
// Summary: Sys the  appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (fi *gcsFileInfo) Sys() interface{} {
	return nil
}
