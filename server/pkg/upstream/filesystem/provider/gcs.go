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
type GcsProvider struct {
	fs     afero.Fs
	client *storage.Client
}

var newStorageClient = storage.NewClient

// NewGcsProvider creates a new GcsProvider from the given configuration.
//
// _ is an unused parameter.
// config holds the configuration settings.
//
// Returns the result.
// Returns an error if the operation fails.
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
// Returns the result.
func (p *GcsProvider) GetFs() afero.Fs {
	return p.fs
}

// ResolvePath resolves the virtual path to a real path in the bucket.
//
// virtualPath is the virtualPath.
//
// Returns the result.
// Returns an error if the operation fails.
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
// Returns an error if the operation fails.
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
// name is the name of the resource.
//
// Returns the result.
// Returns an error if the operation fails.
func (fs *gcsFs) Create(name string) (afero.File, error) {
	return fs.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
}

// Mkdir creates a directory in the filesystem, returning an error, if any happens.
//
// _ is an unused parameter.
// _ is an unused parameter.
//
// Returns an error if the operation fails.
func (fs *gcsFs) Mkdir(_ string, _ os.FileMode) error {
	return nil // Flat namespace
}

// MkdirAll creates a directory path and all parents that does not exist for a given name.
//
// _ is an unused parameter.
// _ is an unused parameter.
//
// Returns an error if the operation fails.
func (fs *gcsFs) MkdirAll(_ string, _ os.FileMode) error {
	return nil // Flat namespace
}

// Open opens a file, returning it or an error, if any happens.
//
// name is the name of the resource.
//
// Returns the result.
// Returns an error if the operation fails.
func (fs *gcsFs) Open(name string) (afero.File, error) {
	return fs.OpenFile(name, os.O_RDONLY, 0)
}

// OpenFile opens a file using the given flags and the given mode.
//
// name is the name of the resource.
// flag is the flag.
// _ is an unused parameter.
//
// Returns the result.
// Returns an error if the operation fails.
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
// name is the name of the resource.
//
// Returns an error if the operation fails.
func (fs *gcsFs) Remove(name string) error {
	return fs.client.Bucket(fs.bucket).Object(name).Delete(fs.ctx)
}

// RemoveAll removes a directory path and any children it contains.
//
// path is the path.
//
// Returns an error if the operation fails.
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
// oldname is the oldname.
// newname is the newname.
//
// Returns an error if the operation fails.
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
// name is the name of the resource.
//
// Returns the result.
// Returns an error if the operation fails.
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
// Returns the result.
func (fs *gcsFs) Name() string {
	return "gcs"
}

// Chmod changes the mode of the named file to mode.
//
// _ is an unused parameter.
// _ is an unused parameter.
//
// Returns an error if the operation fails.
func (fs *gcsFs) Chmod(_ string, _ os.FileMode) error {
	return nil // Not supported
}

// Chown changes the uid and gid of the named file.
//
// _ is an unused parameter.
// _ is an unused parameter.
// _ is an unused parameter.
//
// Returns an error if the operation fails.
func (fs *gcsFs) Chown(_ string, _, _ int) error {
	return nil // Not supported
}

// Chtimes changes the access and modification times of the named file.
//
// _ is an unused parameter.
// _ is an unused parameter.
// _ is an unused parameter.
//
// Returns an error if the operation fails.
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
// Returns an error if the operation fails.
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
// p is the p.
//
// Returns the result.
// Returns an error if the operation fails.
func (f *gcsFile) Read(p []byte) (n int, err error) {
	if f.reader == nil {
		return 0, fmt.Errorf("file not opened for reading")
	}
	return f.reader.Read(p)
}

// ReadAt reads len(b) bytes from the File starting at byte offset off.
//
// p is the p.
// off is the off.
//
// Returns the result.
// Returns an error if the operation fails.
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
// _ is an unused parameter.
// _ is an unused parameter.
//
// Returns the result.
// Returns an error if the operation fails.
func (f *gcsFile) Seek(_ int64, _ int) (int64, error) {
	return 0, fmt.Errorf("seek not supported")
}

// Write writes len(b) bytes to the File.
//
// p is the p.
//
// Returns the result.
// Returns an error if the operation fails.
func (f *gcsFile) Write(p []byte) (n int, err error) {
	if f.writer == nil {
		return 0, fmt.Errorf("file not opened for writing")
	}
	return f.writer.Write(p)
}

// WriteAt writes len(b) bytes to the File starting at byte offset off.
//
// _ is an unused parameter.
// _ is an unused parameter.
//
// Returns the result.
// Returns an error if the operation fails.
func (f *gcsFile) WriteAt(_ []byte, _ int64) (n int, err error) {
	return 0, fmt.Errorf("writeat not supported")
}

// Name returns the name of the file as presented to Open.
//
// Returns the result.
func (f *gcsFile) Name() string {
	return f.name
}

// Readdir reads the contents of the directory associated with file and returns
// a slice of up to n FileInfo values, as would be returned by Lstat, in directory order.
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
// n is the n.
//
// Returns the result.
// Returns an error if the operation fails.
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
// Returns the result.
// Returns an error if the operation fails.
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
// Returns an error if the operation fails.
func (f *gcsFile) Sync() error {
	return nil
}

// Truncate changes the size of the file.
//
// _ is an unused parameter.
//
// Returns an error if the operation fails.
func (f *gcsFile) Truncate(_ int64) error {
	return fmt.Errorf("truncate not supported")
}

// WriteString is like Write, but writes the contents of string s rather than a slice of bytes.
//
// s is the s.
//
// Returns the result.
// Returns an error if the operation fails.
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
// Returns the result.
func (fi *gcsFileInfo) Name() string {
	return fi.name
}

// Size returns the length in bytes for regular files; system-dependent for others.
//
// Returns the result.
func (fi *gcsFileInfo) Size() int64 {
	return fi.size
}

// Mode returns file mode bits.
//
// Returns the result.
func (fi *gcsFileInfo) Mode() os.FileMode {
	if fi.isDir {
		return os.ModeDir | 0755
	}
	return 0644
}

// ModTime returns the modification time.
//
// Returns the result.
func (fi *gcsFileInfo) ModTime() time.Time {
	return fi.modTime
}

// IsDir returns true if the file is a directory.
//
// Returns true if successful.
func (fi *gcsFileInfo) IsDir() bool {
	return fi.isDir
}

// Sys returns underlying data source (can return nil).
//
// Returns the result.
func (fi *gcsFileInfo) Sys() interface{} {
	return nil
}
