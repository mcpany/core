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

// GcsProvider provides access to files in a Google Cloud Storage (GCS) bucket.
//
// Summary: Implements the FilesystemProvider for Google Cloud Storage.
type GcsProvider struct {
	fs     afero.Fs
	client *storage.Client
}

var newStorageClient = storage.NewClient

// NewGcsProvider creates a new GcsProvider from the given configuration.
//
// Summary: Initializes a new GCS filesystem provider and its storage client.
//
// Parameters:
//   - _ (context.Context): Unused context (background context is used for client lifecycle).
//   - config (*configv1.GcsFs): The GCS configuration parameters (bucket name).
//
// Returns:
//   - *GcsProvider: A pointer to the newly created GcsProvider instance.
//   - error: An error if the GCS client cannot be initialized.
//
// Errors:
//   - Returns an error if the configuration object is nil.
//   - Returns an error if the Google Cloud Storage client creation fails.
//
// Side Effects:
//   - Initializes a new Google Cloud Storage client, which may perform authentication and network calls.
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
// Summary: Retrieves the underlying afero GCS filesystem.
//
// Returns:
//   - afero.Fs: An afero.Fs implementation backed by the configured GCS bucket.
//
// Parameters:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (p *GcsProvider) GetFs() afero.Fs {
	return p.fs
}

// ResolvePath resolves the virtual path to a real path in the bucket.
//
// Summary: Resolves and sanitizes a virtual path for use as a GCS object name.
//
// Parameters:
//   - virtualPath (string): The virtual path relative to the bucket root.
//
// Returns:
//   - string: The cleaned and sanitized GCS object name.
//   - error: An error if resolution fails.
//
// Errors:
//   - Returns an error if the resolved path is empty or "." (invalid object name).
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

// Close closes the GCS client and releases resources.
//
// Summary: Closes the underlying Google Cloud Storage client.
//
// Returns:
//   - error: An error if the client fails to close.
//
// Errors:
//   - Returns an error if the storage client's Close() method fails.
//
// Side Effects:
//   - Releases any persistent connections or resources held by the GCS client.
//
// Parameters:
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

// Create creates a new file in the GCS bucket.
//
// Summary: Initializes a create operation for a GCS object.
//
// Parameters:
//   - name (string): The name of the object to create.
//
// Returns:
//   - afero.File: A handle to the newly created GCS object.
//   - error: An error if the creation fails.
//
// Errors:
//   - Returns an error if the underlying OpenFile operation fails.
//
// Side Effects:
//   - None.
func (fs *gcsFs) Create(name string) (afero.File, error) {
	return fs.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
}

// Mkdir is a no-op as GCS uses a flat namespace.
//
// Summary: Simulates directory creation (no-op for GCS).
//
// Parameters:
//   - _ (string): Unused directory name.
//   - _ (os.FileMode): Unused file mode.
//
// Returns:
//   - error: Always nil.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - None.
func (fs *gcsFs) Mkdir(_ string, _ os.FileMode) error {
	return nil // Flat namespace
}

// MkdirAll is a no-op as GCS uses a flat namespace.
//
// Summary: Simulates recursive directory creation (no-op for GCS).
//
// Parameters:
//   - _ (string): Unused path.
//   - _ (os.FileMode): Unused file mode.
//
// Returns:
//   - error: Always nil.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - None.
func (fs *gcsFs) MkdirAll(_ string, _ os.FileMode) error {
	return nil // Flat namespace
}

// Open opens an existing GCS object for reading.
//
// Summary: Opens a GCS object for read access.
//
// Parameters:
//   - name (string): The name of the GCS object.
//
// Returns:
//   - afero.File: A handle to the opened GCS object.
//   - error: An error if the object does not exist or opening fails.
//
// Errors:
//   - Returns an error if the object does not exist.
//   - Returns an error if the storage client fails to initiate a reader.
//
// Side Effects:
//   - None.
func (fs *gcsFs) Open(name string) (afero.File, error) {
	return fs.OpenFile(name, os.O_RDONLY, 0)
}

// OpenFile opens a GCS object using specific flags.
//
// Summary: Opens a GCS object with the specified access flags.
//
// Parameters:
//   - name (string): The name of the GCS object.
//   - flag (int): Standard OS file flags (O_RDONLY, O_WRONLY, O_RDWR, etc.).
//   - _ (os.FileMode): Unused file mode (GCS uses bucket-level ACLs).
//
// Returns:
//   - afero.File: A handle to the GCS object.
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns os.ErrNotExist if the object is missing during a read operation.
//   - Returns an error if the storage client fails to create a reader or writer.
//
// Side Effects:
//   - Initiates a network connection to GCS for reading or writing.
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

// Remove deletes an object from the GCS bucket.
//
// Summary: Deletes a specific GCS object.
//
// Parameters:
//   - name (string): The name of the object to delete.
//
// Returns:
//   - error: An error if the deletion fails.
//
// Errors:
//   - Returns an error if the storage client fails to delete the object.
//
// Side Effects:
//   - Permanently deletes data from the GCS bucket.
func (fs *gcsFs) Remove(name string) error {
	return fs.client.Bucket(fs.bucket).Object(name).Delete(fs.ctx)
}

// RemoveAll deletes all objects with the given prefix.
//
// Summary: Performs a recursive deletion by prefix in the GCS bucket.
//
// Parameters:
//   - path (string): The prefix for objects to be deleted.
//
// Returns:
//   - error: An error if any deletion fails or the object iterator fails.
//
// Errors:
//   - Returns an error if the storage client fails to list or delete objects.
//
// Side Effects:
//   - Iteratively deletes all matching objects from the bucket.
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

// Rename copies an object to a new name and deletes the original.
//
// Summary: Renames a GCS object via a copy-and-delete sequence.
//
// Parameters:
//   - oldname (string): The current object name.
//   - newname (string): The new object name.
//
// Returns:
//   - error: An error if the rename operation fails.
//
// Errors:
//   - Returns an error if the copy operation fails.
//   - Returns an error if the deletion of the original object fails.
//
// Side Effects:
//   - Creates a new object and deletes the old one in the GCS bucket.
func (fs *gcsFs) Rename(oldname, newname string) error {
	src := fs.client.Bucket(fs.bucket).Object(oldname)
	dst := fs.client.Bucket(fs.bucket).Object(newname)

	if _, err := dst.CopierFrom(src).Run(fs.ctx); err != nil {
		return err
	}
	return src.Delete(fs.ctx)
}

// Stat retrieves metadata for a GCS object.
//
// Summary: Fetches file information for a GCS object.
//
// Parameters:
//   - name (string): The name of the object to stat.
//
// Returns:
//   - os.FileInfo: Metadata for the GCS object.
//   - error: An error if the object does not exist or metadata cannot be fetched.
//
// Errors:
//   - Returns os.ErrNotExist if the object is missing.
//   - Returns an error if the storage client fails to fetch attributes.
//
// Side Effects:
//   - Makes a network call to fetch object attributes.
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

// Name returns the identifier for this filesystem implementation.
//
// Summary: Retrieves the filesystem name.
//
// Returns:
//   - string: The constant string "gcs".
//
// Parameters:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (fs *gcsFs) Name() string {
	return "gcs"
}

// Chmod is not supported by GCS.
//
// Summary: Changes file mode (unsupported for GCS).
//
// Parameters:
//   - _ (string): Unused object name.
//   - _ (os.FileMode): Unused mode.
//
// Returns:
//   - error: Nil, as GCS doesn't support traditional file modes.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - None.
func (fs *gcsFs) Chmod(_ string, _ os.FileMode) error {
	return nil // Not supported
}

// Chown is not supported by GCS.
//
// Summary: Changes object ownership (unsupported for GCS).
//
// Parameters:
//   - _ (string): Unused object name.
//   - _ (int): Unused uid.
//   - _ (int): Unused gid.
//
// Returns:
//   - error: Nil, as GCS doesn't support UID/GID ownership.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - None.
func (fs *gcsFs) Chown(_ string, _, _ int) error {
	return nil // Not supported
}

// Chtimes is not supported by GCS.
//
// Summary: Changes object timestamps (unsupported for GCS).
//
// Parameters:
//   - _ (string): Unused object name.
//   - _ (time.Time): Unused atime.
//   - _ (time.Time): Unused mtime.
//
// Returns:
//   - error: Nil, as GCS manages its own timestamps.
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

// Close closes the GCS file reader or writer.
//
// Summary: Releases resources associated with an open GCS object.
//
// Returns:
//   - error: An error if the underlying reader or writer fails to close.
//
// Errors:
//   - Returns an error if the writer.Close() fails (e.g. during upload flush).
//
// Side Effects:
//   - Flushes remaining data to GCS if the file was opened for writing.
//
// Parameters:
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

// Read reads data from the GCS object into the provided buffer.
//
// Summary: Reads bytes from an open GCS object.
//
// Parameters:
//   - p ([]byte): The destination buffer for the read data.
//
// Returns:
//   - n (int): The number of bytes successfully read.
//   - err (error): An error if the read operation fails or EOF is reached.
//
// Errors:
//   - Returns an error if the file was not opened for reading.
//   - Returns an error from the underlying GCS reader.
//
// Side Effects:
//   - Performs a network read operation from GCS.
func (f *gcsFile) Read(p []byte) (n int, err error) {
	if f.reader == nil {
		return 0, fmt.Errorf("file not opened for reading")
	}
	return f.reader.Read(p)
}

// ReadAt reads a specific byte range from the GCS object.
//
// Summary: Performs a positional read from a GCS object.
//
// Parameters:
//   - p ([]byte): The destination buffer for the read data.
//   - off (int64): The byte offset to start reading from.
//
// Returns:
//   - n (int): The number of bytes successfully read.
//   - err (error): An error if the read operation fails.
//
// Errors:
//   - Returns an error if the storage client fails to create a range reader.
//   - Returns io.ErrUnexpectedEOF if the read finishes before the buffer is full.
//
// Side Effects:
//   - Initiates a new network request for the specific byte range.
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

// Seek is not supported for GCS objects.
//
// Summary: Changes the current file offset (unsupported for GCS).
//
// Parameters:
//   - _ (int64): Unused offset.
//   - _ (int): Unused whence.
//
// Returns:
//   - int64: Always 0.
//   - error: An error indicating that seek is not supported.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - None.
func (f *gcsFile) Seek(_ int64, _ int) (int64, error) {
	return 0, fmt.Errorf("seek not supported")
}

// Write sends data to the GCS object writer.
//
// Summary: Writes bytes to a GCS object.
//
// Parameters:
//   - p ([]byte): The data buffer to be written.
//
// Returns:
//   - n (int): The number of bytes successfully written.
//   - err (error): An error if the write operation fails.
//
// Errors:
//   - Returns an error if the object was not opened for writing.
//   - Returns an error from the underlying storage writer.
//
// Side Effects:
//   - Performs network writing to GCS (buffered).
func (f *gcsFile) Write(p []byte) (n int, err error) {
	if f.writer == nil {
		return 0, fmt.Errorf("file not opened for writing")
	}
	return f.writer.Write(p)
}

// WriteAt is not supported for GCS objects.
//
// Summary: Performs a positional write (unsupported for GCS).
//
// Parameters:
//   - _ ([]byte): Unused data.
//   - _ (int64): Unused offset.
//
// Returns:
//   - n (int): Always 0.
//   - err (error): An error indicating that WriteAt is not supported.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - None.
func (f *gcsFile) WriteAt(_ []byte, _ int64) (n int, err error) {
	return 0, fmt.Errorf("writeat not supported")
}

// Name returns the identifier for the open GCS object.
//
// Summary: Retrieves the name of the file.
//
// Returns:
//   - string: The name of the GCS object.
//
// Parameters:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (f *gcsFile) Name() string {
	return f.name
}

// Readdir lists objects under the current "directory" prefix.
//
// Summary: Retrieves a list of object metadata under the current path.
//
// Parameters:
//   - _ (int): Unused limit count (lists all matching objects).
//
// Returns:
//   - []os.FileInfo: A slice of file information for objects and sub-prefixes.
//   - error: An error if the object iterator fails.
//
// Side Effects:
//   - Performs multiple network calls to iterate through objects in the bucket.
//
// Errors:
//   - Returns an error if the operation fails.
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

// Readdirnames returns a slice of object names in the current "directory".
//
// Summary: Retrieves a list of object names under the current path.
//
// Parameters:
//   - n (int): Maximum number of names to return.
//
// Returns:
//   - []string: A slice of names for objects and sub-prefixes.
//   - error: An error if Readdir fails.
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

// Stat returns metadata for the open GCS object.
//
// Summary: Fetches file information for an open file handle.
//
// Returns:
//   - os.FileInfo: Metadata for the current object.
//   - error: An error if the stat operation fails.
//
// Side Effects:
//   - May make a network call if attributes aren't cached in the handle.
//
// Parameters:
//   - None.
//
// Errors:
//   - Returns an error if the operation fails.
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

// Sync is a no-op for GCS objects.
//
// Summary: Flushes data to stable storage (no-op for GCS handle).
//
// Returns:
//   - error: Always nil.
//
// Parameters:
//   - None.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - None.
func (f *gcsFile) Sync() error {
	return nil
}

// Truncate is not supported for GCS objects.
//
// Summary: Changes the size of the GCS object (unsupported).
//
// Parameters:
//   - _ (int64): Unused size target.
//
// Returns:
//   - error: An error indicating that truncate is not supported.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - None.
func (f *gcsFile) Truncate(_ int64) error {
	return fmt.Errorf("truncate not supported")
}

// WriteString writes the provided string to the GCS object.
//
// Summary: Writes a string to an open GCS object writer.
//
// Parameters:
//   - s (string): The string to write.
//
// Returns:
//   - ret (int): The number of bytes written.
//   - err (error): An error if the write operation fails.
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

// Name returns the base name of the object.
//
// Summary: Retrieves the file name from FileInfo.
//
// Returns:
//   - string: The object name.
//
// Parameters:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (fi *gcsFileInfo) Name() string {
	return fi.name
}

// Size returns the size of the GCS object in bytes.
//
// Summary: Retrieves the object size.
//
// Returns:
//   - int64: The object size in bytes.
//
// Parameters:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (fi *gcsFileInfo) Size() int64 {
	return fi.size
}

// Mode returns a simulated file mode for the GCS object.
//
// Summary: Retrieves a simulated file mode.
//
// Returns:
//   - os.FileMode: ModeDir | 0755 for directories, or 0644 for regular objects.
//
// Parameters:
//   - None.
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

// ModTime returns the time the GCS object was last updated.
//
// Summary: Retrieves the object modification time.
//
// Returns:
//   - time.Time: The last updated timestamp.
//
// Parameters:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (fi *gcsFileInfo) ModTime() time.Time {
	return fi.modTime
}

// IsDir returns true if this FileInfo represents a virtual directory (prefix).
//
// Summary: Checks if the object is a directory.
//
// Returns:
//   - bool: True if this represents a directory prefix.
//
// Parameters:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (fi *gcsFileInfo) IsDir() bool {
	return fi.isDir
}

// Sys returns the underlying data source, which is always nil for GCS FileInfo.
//
// Summary: Retrieves internal system details (unsupported).
//
// Returns:
//   - any: Always nil.
//
// Parameters:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (fi *gcsFileInfo) Sys() interface{} {
	return nil
}
