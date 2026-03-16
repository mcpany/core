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
// Summary: GcsProvider provides access to files in a Google Cloud Storage bucket.
//
// Summary: GcsProvider provides access to files in a Google Cloud Storage bucket.
type GcsProvider struct {
	fs     afero.Fs
	client *storage.Client
}

// NewGcsProvider creates a new GcsProvider from the given configuration.
//
// Summary: NewGcsProvider creates a new GcsProvider from the given configuration.
//
// Parameters:
//   - _ (context.Context): The provided _ data.
//   - config (*configv1.GcsFs): The configuration settings.
//
// Returns:
//   - *GcsProvider: The resulting object or data structure.
//   - error: An error if the execution fails, otherwise nil.
//
// Errors:
//   - Returns an error if the operation fails, invalid input is provided, or a downstream dependency fails.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
//   - Returns an error if the operation fails, invalid input is provided, or a downstream dependency fails.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
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
// GetFs returns the underlying filesystem.
//
// Summary: GetFs returns the underlying filesystem.
//
// Parameters:
//   - None.
//
// Returns:
//   - afero.Fs: The resulting object or data structure.
//
// Errors:
//   - None.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
//
// Parameters:
//   - None.
//
// ResolvePath resolves the virtual path to a real path in the bucket.
//
// Summary: ResolvePath resolves the virtual path to a real path in the bucket.
//
// Parameters:
//   - virtualPath (string): The textual representation of virtualpath.
//
// Returns:
//   - string: The resulting text.
//   - error: An error if the execution fails, otherwise nil.
//
// Errors:
//   - Returns an error if the operation fails, invalid input is provided, or a downstream dependency fails.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
// Summary: ResolvePath resolves the virtual path to a real path in the bucket.
//
// Parameters:
//   - virtualPath (string): The textual representation of virtualpath.
//
// Returns:
//   - string: The resulting text.
//   - error: An error if the execution fails, otherwise nil.
//
// Errors:
//   - Returns an error if the operation fails, invalid input is provided, or a downstream dependency fails.
// Close closes the GCS client.
//
// Summary: Close closes the GCS client.
//
// Parameters:
//   - None.
//
// Returns:
//   - error: An error if the execution fails, otherwise nil.
//
// Errors:
//   - Returns an error if the operation fails, invalid input is provided, or a downstream dependency fails.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
	}
	return cleanPath, nil
}

// Close closes the GCS client.
//
// Summary: Close closes the GCS client.
//
// Parameters:
//   - None.
//
// Returns:
//   - error: An error if the execution fails, otherwise nil.
//
// Errors:
// Create creates a file in the filesystem, returning the file and an error, if any happens.
//
// Summary: Create creates a file in the filesystem, returning the file and an error, if any happens.
//
// Parameters:
//   - name (string): The human-readable or system name.
//
// Returns:
//   - afero.File: The resulting object or data structure.
//   - error: An error if the execution fails, otherwise nil.
//
// Errors:
//   - Returns an error if the operation fails, invalid input is provided, or a downstream dependency fails.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
	client *storage.Client
	bucket string
	ctx    context.Context
}
// Mkdir creates a directory in the filesystem, returning an error, if any happens.
//
// Summary: Mkdir creates a directory in the filesystem, returning an error, if any happens.
//
// Parameters:
//   - _ (string): The textual representation of _.
//   - _ (os.FileMode): The provided _ data.
//
// Returns:
//   - error: An error if the execution fails, otherwise nil.
//
// Errors:
//   - Returns an error if the operation fails, invalid input is provided, or a downstream dependency fails.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
func (fs *gcsFs) Create(name string) (afero.File, error) {
// MkdirAll creates a directory path and all parents that does not exist for a given name.
//
// Summary: MkdirAll creates a directory path and all parents that does not exist for a given name.
//
// Parameters:
//   - _ (string): The textual representation of _.
//   - _ (os.FileMode): The provided _ data.
//
// Returns:
//   - error: An error if the execution fails, otherwise nil.
//
// Errors:
//   - Returns an error if the operation fails, invalid input is provided, or a downstream dependency fails.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
// Errors:
//   - Returns an error if the operation fails, invalid input is provided, or a downstream dependency fails.
//
// Side Effects:
// Open opens a file, returning it or an error, if any happens.
//
// Summary: Open opens a file, returning it or an error, if any happens.
//
// Parameters:
//   - name (string): The human-readable or system name.
//
// Returns:
//   - afero.File: The resulting object or data structure.
//   - error: An error if the execution fails, otherwise nil.
//
// Errors:
//   - Returns an error if the operation fails, invalid input is provided, or a downstream dependency fails.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
//   - error: An error if the execution fails, otherwise nil.
//
// Errors:
//   - Returns an error if the operation fails, invalid input is provided, or a downstream dependency fails.
// OpenFile opens a file using the given flags and the given mode.
//
// Summary: OpenFile opens a file using the given flags and the given mode.
//
// Parameters:
//   - name (string): The human-readable or system name.
//   - flag (int): The numeric value for flag.
//   - _ (os.FileMode): The provided _ data.
//
// Returns:
//   - afero.File: The resulting object or data structure.
//   - error: An error if the execution fails, otherwise nil.
//
// Errors:
//   - Returns an error if the operation fails, invalid input is provided, or a downstream dependency fails.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
//   - error: An error if the execution fails, otherwise nil.
//
// Errors:
//   - Returns an error if the operation fails, invalid input is provided, or a downstream dependency fails.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
func (fs *gcsFs) Open(name string) (afero.File, error) {
	return fs.OpenFile(name, os.O_RDONLY, 0)
}

// OpenFile opens a file using the given flags and the given mode.
//
// Summary: OpenFile opens a file using the given flags and the given mode.
//
// Parameters:
//   - name (string): The human-readable or system name.
//   - flag (int): The numeric value for flag.
//   - _ (os.FileMode): The provided _ data.
//
// Returns:
//   - afero.File: The resulting object or data structure.
//   - error: An error if the execution fails, otherwise nil.
//
// Errors:
// Remove removes a file identified by name, returning an error, if any happens.
//
// Summary: Remove removes a file identified by name, returning an error, if any happens.
//
// Parameters:
//   - name (string): The human-readable or system name.
//
// Returns:
//   - error: An error if the execution fails, otherwise nil.
//
// Errors:
//   - Returns an error if the operation fails, invalid input is provided, or a downstream dependency fails.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
		f.writer = wc
		return f, nil
	}

// RemoveAll removes a directory path and any children it contains.
//
// Summary: RemoveAll removes a directory path and any children it contains.
//
// Parameters:
//   - path (string): The textual representation of path.
//
// Returns:
//   - error: An error if the execution fails, otherwise nil.
//
// Errors:
//   - Returns an error if the operation fails, invalid input is provided, or a downstream dependency fails.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
//
// Summary: Remove removes a file identified by name, returning an error, if any happens.
//
// Parameters:
//   - name (string): The human-readable or system name.
//
// Returns:
//   - error: An error if the execution fails, otherwise nil.
//
// Errors:
//   - Returns an error if the operation fails, invalid input is provided, or a downstream dependency fails.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
func (fs *gcsFs) Remove(name string) error {
	return fs.client.Bucket(fs.bucket).Object(name).Delete(fs.ctx)
}

// Rename renames a file.
//
// Summary: Rename renames a file.
//
// Parameters:
//   - oldname (string): The human-readable or system name.
//   - newname (string): The human-readable or system name.
//
// Returns:
//   - error: An error if the execution fails, otherwise nil.
//
// Errors:
//   - Returns an error if the operation fails, invalid input is provided, or a downstream dependency fails.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
//   - May modify internal state or perform external network calls.
func (fs *gcsFs) RemoveAll(path string) error {
	// Delete everything with prefix
	it := fs.client.Bucket(fs.bucket).Objects(fs.ctx, &storage.Query{Prefix: path})
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
// Stat returns a FileInfo describing the named file, or an error, if any happens.
//
// Summary: Stat returns a FileInfo describing the named file, or an error, if any happens.
//
// Parameters:
//   - name (string): The human-readable or system name.
//
// Returns:
//   - os.FileInfo: The resulting object or data structure.
//   - error: An error if the execution fails, otherwise nil.
//
// Errors:
//   - Returns an error if the operation fails, invalid input is provided, or a downstream dependency fails.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
//   - oldname (string): The human-readable or system name.
//   - newname (string): The human-readable or system name.
//
// Returns:
//   - error: An error if the execution fails, otherwise nil.
//
// Errors:
//   - Returns an error if the operation fails, invalid input is provided, or a downstream dependency fails.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
func (fs *gcsFs) Rename(oldname, newname string) error {
	src := fs.client.Bucket(fs.bucket).Object(oldname)
	dst := fs.client.Bucket(fs.bucket).Object(newname)

	if _, err := dst.CopierFrom(src).Run(fs.ctx); err != nil {
// Name returns the name of this file system.
//
// Summary: Name returns the name of this file system.
//
// Parameters:
//   - None.
//
// Returns:
//   - string: The resulting text.
//
// Errors:
//   - None.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
// Summary: Stat returns a FileInfo describing the named file, or an error, if any happens.
//
// Parameters:
//   - name (string): The human-readable or system name.
// Chmod changes the mode of the named file to mode.
//
// Summary: Chmod changes the mode of the named file to mode.
//
// Parameters:
//   - _ (string): The textual representation of _.
//   - _ (os.FileMode): The provided _ data.
//
// Returns:
//   - error: An error if the execution fails, otherwise nil.
//
// Errors:
//   - Returns an error if the operation fails, invalid input is provided, or a downstream dependency fails.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
			return nil, os.ErrNotExist
		}
		return nil, err
	}
// Chown changes the uid and gid of the named file.
//
// Summary: Chown changes the uid and gid of the named file.
//
// Parameters:
//   - _ (string): The textual representation of _.
//   - _ (int): The numeric value for _.
//   - _ (int): The numeric value for _.
//
// Returns:
//   - error: An error if the execution fails, otherwise nil.
//
// Errors:
//   - Returns an error if the operation fails, invalid input is provided, or a downstream dependency fails.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
// Returns:
//   - string: The resulting text.
//
// Errors:
// Chtimes changes the access and modification times of the named file.
//
// Summary: Chtimes changes the access and modification times of the named file.
//
// Parameters:
//   - _ (string): The textual representation of _.
//   - _ (time.Time): The provided _ data.
//   - _ (time.Time): The provided _ data.
//
// Returns:
//   - error: An error if the execution fails, otherwise nil.
//
// Errors:
//   - Returns an error if the operation fails, invalid input is provided, or a downstream dependency fails.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
//
// Returns:
//   - error: An error if the execution fails, otherwise nil.
//
// Errors:
//   - Returns an error if the operation fails, invalid input is provided, or a downstream dependency fails.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
func (fs *gcsFs) Chmod(_ string, _ os.FileMode) error {
	return nil // Not supported
// Close closes the file.
//
// Summary: Close closes the file.
//
// Parameters:
//   - None.
//
// Returns:
//   - error: An error if the execution fails, otherwise nil.
//
// Errors:
//   - Returns an error if the operation fails, invalid input is provided, or a downstream dependency fails.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
//
// Returns:
//   - error: An error if the execution fails, otherwise nil.
//
// Errors:
//   - Returns an error if the operation fails, invalid input is provided, or a downstream dependency fails.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
func (fs *gcsFs) Chown(_ string, _, _ int) error {
// Read reads up to len(b) bytes from the File.
//
// Summary: Read reads up to len(b) bytes from the File.
//
// Parameters:
//   - p ([]byte): The provided p data.
//
// Returns:
//   - n (int): The calculated numeric value.
//   - err (error): An error if the execution fails, otherwise nil.
//
// Errors:
//   - Returns an error if the operation fails, invalid input is provided, or a downstream dependency fails.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
//
// Errors:
//   - Returns an error if the operation fails, invalid input is provided, or a downstream dependency fails.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
func (fs *gcsFs) Chtimes(_ string, _, _ time.Time) error {
// ReadAt reads len(b) bytes from the File starting at byte offset off.
//
// Summary: ReadAt reads len(b) bytes from the File starting at byte offset off.
//
// Parameters:
//   - p ([]byte): The provided p data.
//   - off (int64): The numeric value for off.
//
// Returns:
//   - n (int): The calculated numeric value.
//   - err (error): An error if the execution fails, otherwise nil.
//
// Errors:
//   - Returns an error if the operation fails, invalid input is provided, or a downstream dependency fails.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
//   - None.
//
// Returns:
//   - error: An error if the execution fails, otherwise nil.
//
// Errors:
//   - Returns an error if the operation fails, invalid input is provided, or a downstream dependency fails.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
func (f *gcsFile) Close() error {
	if f.writer != nil {
// Seek sets the offset for the next Read or Write to offset, interpreted according to whence.
//
// Summary: Seek sets the offset for the next Read or Write to offset, interpreted according to whence.
//
// Parameters:
//   - _ (int64): The numeric value for _.
//   - _ (int): The numeric value for _.
//
// Returns:
//   - int64: The calculated numeric value.
//   - error: An error if the execution fails, otherwise nil.
//
// Errors:
//   - Returns an error if the operation fails, invalid input is provided, or a downstream dependency fails.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
// Returns:
//   - n (int): The calculated numeric value.
//   - err (error): An error if the execution fails, otherwise nil.
//
// Write writes len(b) bytes to the File.
//
// Summary: Write writes len(b) bytes to the File.
//
// Parameters:
//   - p ([]byte): The provided p data.
//
// Returns:
//   - n (int): The calculated numeric value.
//   - err (error): An error if the execution fails, otherwise nil.
//
// Errors:
//   - Returns an error if the operation fails, invalid input is provided, or a downstream dependency fails.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
// Summary: ReadAt reads len(b) bytes from the File starting at byte offset off.
//
// Parameters:
//   - p ([]byte): The provided p data.
//   - off (int64): The numeric value for off.
//
// Returns:
// WriteAt writes len(b) bytes to the File starting at byte offset off.
//
// Summary: WriteAt writes len(b) bytes to the File starting at byte offset off.
//
// Parameters:
//   - _ ([]byte): The provided _ data.
//   - _ (int64): The numeric value for _.
//
// Returns:
//   - n (int): The calculated numeric value.
//   - err (error): An error if the execution fails, otherwise nil.
//
// Errors:
//   - Returns an error if the operation fails, invalid input is provided, or a downstream dependency fails.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
	}
	defer func() { _ = rc.Close() }()
	return io.ReadFull(rc, p)
}
// Name returns the name of the file as presented to Open.
//
// Summary: Name returns the name of the file as presented to Open.
//
// Parameters:
//   - None.
//
// Returns:
//   - string: The resulting text.
//
// Errors:
//   - None.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
//   - _ (int): The numeric value for _.
//
// Returns:
//   - int64: The calculated numeric value.
// Readdir reads the contents of the directory associated with file and returns
//
// Summary: Readdir reads the contents of the directory associated with file and returns
//
// Parameters:
//   - _ (int): The numeric value for _.
//
// Returns:
//   - []os.FileInfo: The resulting object or data structure.
//   - error: An error if the execution fails, otherwise nil.
//
// Errors:
//   - Returns an error if the operation fails, invalid input is provided, or a downstream dependency fails.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
// Parameters:
//   - p ([]byte): The provided p data.
//
// Returns:
//   - n (int): The calculated numeric value.
//   - err (error): An error if the execution fails, otherwise nil.
//
// Errors:
//   - Returns an error if the operation fails, invalid input is provided, or a downstream dependency fails.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
func (f *gcsFile) Write(p []byte) (n int, err error) {
	if f.writer == nil {
		return 0, fmt.Errorf("file not opened for writing")
	}
	return f.writer.Write(p)
}

// WriteAt writes len(b) bytes to the File starting at byte offset off.
//
// Summary: WriteAt writes len(b) bytes to the File starting at byte offset off.
//
// Parameters:
//   - _ ([]byte): The provided _ data.
//   - _ (int64): The numeric value for _.
//
// Returns:
//   - n (int): The calculated numeric value.
//   - err (error): An error if the execution fails, otherwise nil.
//
// Errors:
//   - Returns an error if the operation fails, invalid input is provided, or a downstream dependency fails.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
func (f *gcsFile) WriteAt(_ []byte, _ int64) (n int, err error) {
	return 0, fmt.Errorf("writeat not supported")
}

// Name returns the name of the file as presented to Open.
//
// Summary: Name returns the name of the file as presented to Open.
// Readdirnames reads and returns a slice of names from the directory f.
//
// Summary: Readdirnames reads and returns a slice of names from the directory f.
//
// Parameters:
//   - n (int): The numeric value for n.
//
// Returns:
//   - []string: The resulting text.
//   - error: An error if the execution fails, otherwise nil.
//
// Errors:
//   - Returns an error if the operation fails, invalid input is provided, or a downstream dependency fails.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
}

// Readdir reads the contents of the directory associated with file and returns
//
// Summary: Readdir reads the contents of the directory associated with file and returns
//
// Parameters:
//   - _ (int): The numeric value for _.
//
// Returns:
//   - []os.FileInfo: The resulting object or data structure.
//   - error: An error if the execution fails, otherwise nil.
// Stat returns the FileInfo structure describing file.
//
// Summary: Stat returns the FileInfo structure describing file.
//
// Parameters:
//   - None.
//
// Returns:
//   - os.FileInfo: The resulting object or data structure.
//   - error: An error if the execution fails, otherwise nil.
//
// Errors:
//   - Returns an error if the operation fails, invalid input is provided, or a downstream dependency fails.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
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
// Sync commits the current contents of the file to stable storage.
//
// Summary: Sync commits the current contents of the file to stable storage.
//
// Parameters:
//   - None.
//
// Returns:
//   - error: An error if the execution fails, otherwise nil.
//
// Errors:
//   - Returns an error if the operation fails, invalid input is provided, or a downstream dependency fails.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
				isDir:   false,
			})
		}
	}
// Truncate changes the size of the file.
//
// Summary: Truncate changes the size of the file.
//
// Parameters:
//   - _ (int64): The numeric value for _.
//
// Returns:
//   - error: An error if the execution fails, otherwise nil.
//
// Errors:
//   - Returns an error if the operation fails, invalid input is provided, or a downstream dependency fails.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
//
// Errors:
//   - Returns an error if the operation fails, invalid input is provided, or a downstream dependency fails.
//
// WriteString is like Write, but writes the contents of string s rather than a slice of bytes.
//
// Summary: WriteString is like Write, but writes the contents of string s rather than a slice of bytes.
//
// Parameters:
//   - s (string): The textual representation of s.
//
// Returns:
//   - ret (int): The calculated numeric value.
//   - err (error): An error if the execution fails, otherwise nil.
//
// Errors:
//   - Returns an error if the operation fails, invalid input is provided, or a downstream dependency fails.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
// Stat returns the FileInfo structure describing file.
//
// Summary: Stat returns the FileInfo structure describing file.
//
// Parameters:
//   - None.
//
// Returns:
//   - os.FileInfo: The resulting object or data structure.
//   - error: An error if the execution fails, otherwise nil.
//
// Name returns the base name of the file.
//
// Summary: Name returns the base name of the file.
//
// Parameters:
//   - None.
//
// Returns:
//   - string: The resulting text.
//
// Errors:
//   - None.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
		return &gcsFileInfo{
			name:    f.name, // ReaderObjectAttrs doesn't always have name?
			size:    f.reader.Attrs.Size,
			modTime: f.reader.Attrs.LastModified,
// Size returns the length in bytes for regular files; system-dependent for others.
//
// Summary: Size returns the length in bytes for regular files; system-dependent for others.
//
// Parameters:
//   - None.
//
// Returns:
//   - int64: The calculated numeric value.
//
// Errors:
//   - None.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
			size:  0,
			isDir: false,
		}, nil
	}
// Mode returns file mode bits.
//
// Summary: Mode returns file mode bits.
//
// Parameters:
//   - None.
//
// Returns:
//   - os.FileMode: The resulting object or data structure.
//
// Errors:
//   - None.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
//
// Parameters:
//   - None.
//
// Returns:
//   - error: An error if the execution fails, otherwise nil.
//
// ModTime returns the modification time.
//
// Summary: ModTime returns the modification time.
//
// Parameters:
//   - None.
//
// Returns:
//   - time.Time: The resulting object or data structure.
//
// Errors:
//   - None.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
}

// Truncate changes the size of the file.
//
// IsDir returns true if the file is a directory.
//
// Summary: IsDir returns true if the file is a directory.
//
// Parameters:
//   - None.
//
// Returns:
//   - bool: True if successful or valid, false otherwise.
//
// Errors:
//   - None.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
//
// Errors:
//   - Returns an error if the operation fails, invalid input is provided, or a downstream dependency fails.
//
// Sys returns underlying data source (can return nil).
//
// Summary: Sys returns underlying data source (can return nil).
//
// Parameters:
//   - None.
//
// Returns:
//   - interface{}: The calculated numeric value.
//
// Errors:
//   - None.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
//
// Summary: WriteString is like Write, but writes the contents of string s rather than a slice of bytes.
//
// Parameters:
//   - s (string): The textual representation of s.
//
// Returns:
//   - ret (int): The calculated numeric value.
//   - err (error): An error if the execution fails, otherwise nil.
//
// Errors:
//   - Returns an error if the operation fails, invalid input is provided, or a downstream dependency fails.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
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
// Summary: Name returns the base name of the file.
//
// Parameters:
//   - None.
//
// Returns:
//   - string: The resulting text.
//
// Errors:
//   - None.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
func (fi *gcsFileInfo) Name() string {
	return fi.name
}

// Size returns the length in bytes for regular files; system-dependent for others.
//
// Summary: Size returns the length in bytes for regular files; system-dependent for others.
//
// Parameters:
//   - None.
//
// Returns:
//   - int64: The calculated numeric value.
//
// Errors:
//   - None.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
func (fi *gcsFileInfo) Size() int64 {
	return fi.size
}

// Mode returns file mode bits.
//
// Summary: Mode returns file mode bits.
//
// Parameters:
//   - None.
//
// Returns:
//   - os.FileMode: The resulting object or data structure.
//
// Errors:
//   - None.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
func (fi *gcsFileInfo) Mode() os.FileMode {
	if fi.isDir {
		return os.ModeDir | 0755
	}
	return 0644
}

// ModTime returns the modification time.
//
// Summary: ModTime returns the modification time.
//
// Parameters:
//   - None.
//
// Returns:
//   - time.Time: The resulting object or data structure.
//
// Errors:
//   - None.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
func (fi *gcsFileInfo) ModTime() time.Time {
	return fi.modTime
}

// IsDir returns true if the file is a directory.
//
// Summary: IsDir returns true if the file is a directory.
//
// Parameters:
//   - None.
//
// Returns:
//   - bool: True if successful or valid, false otherwise.
//
// Errors:
//   - None.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
func (fi *gcsFileInfo) IsDir() bool {
	return fi.isDir
}

// Sys returns underlying data source (can return nil).
//
// Summary: Sys returns underlying data source (can return nil).
//
// Parameters:
//   - None.
//
// Returns:
//   - interface{}: The calculated numeric value.
//
// Errors:
//   - None.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
func (fi *gcsFileInfo) Sys() interface{} {
	return nil
}
