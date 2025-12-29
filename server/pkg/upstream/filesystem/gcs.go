// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package filesystem

import (
	"context"
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

func (u *Upstream) createGcsFilesystem(ctx context.Context, config *configv1.GcsFs) (afero.Fs, error) {
	if config == nil {
		return nil, fmt.Errorf("gcs config is nil")
	}

	client, err := storage.NewClient(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to create gcs client: %w", err)
	}

	u.mu.Lock()
	u.closers = append(u.closers, client)
	u.mu.Unlock()

	return &gcsFs{client: client, bucket: config.GetBucket(), ctx: context.Background()}, nil
}

type gcsFs struct {
	client *storage.Client
	bucket string
	ctx    context.Context
}

func (fs *gcsFs) Create(name string) (afero.File, error) {
	return fs.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
}

func (fs *gcsFs) Mkdir(name string, perm os.FileMode) error {
	return nil // Flat namespace
}

func (fs *gcsFs) MkdirAll(path string, perm os.FileMode) error {
	return nil // Flat namespace
}

func (fs *gcsFs) Open(name string) (afero.File, error) {
	return fs.OpenFile(name, os.O_RDONLY, 0)
}

func (fs *gcsFs) OpenFile(name string, flag int, perm os.FileMode) (afero.File, error) {
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
		if err == storage.ErrObjectNotExist {
			return nil, os.ErrNotExist
		}
		return nil, err
	}
	f.reader = rc
	return f, nil
}

func (fs *gcsFs) Remove(name string) error {
	return fs.client.Bucket(fs.bucket).Object(name).Delete(fs.ctx)
}

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

func (fs *gcsFs) Rename(oldname, newname string) error {
	src := fs.client.Bucket(fs.bucket).Object(oldname)
	dst := fs.client.Bucket(fs.bucket).Object(newname)

	if _, err := dst.CopierFrom(src).Run(fs.ctx); err != nil {
		return err
	}
	return src.Delete(fs.ctx)
}

func (fs *gcsFs) Stat(name string) (os.FileInfo, error) {
	attrs, err := fs.client.Bucket(fs.bucket).Object(name).Attrs(fs.ctx)
	if err != nil {
		if err == storage.ErrObjectNotExist {
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

func (fs *gcsFs) Name() string {
	return "gcs"
}

func (fs *gcsFs) Chmod(name string, mode os.FileMode) error {
	return nil // Not supported
}

func (fs *gcsFs) Chown(name string, uid, gid int) error {
	return nil // Not supported
}

func (fs *gcsFs) Chtimes(name string, atime time.Time, mtime time.Time) error {
	return nil // Not supported
}

type gcsFile struct {
	fs     *gcsFs
	name   string
	reader *storage.Reader
	writer *storage.Writer
}

func (f *gcsFile) Close() error {
	if f.writer != nil {
		return f.writer.Close()
	}
	if f.reader != nil {
		return f.reader.Close()
	}
	return nil
}

func (f *gcsFile) Read(p []byte) (n int, err error) {
	if f.reader == nil {
		return 0, fmt.Errorf("file not opened for reading")
	}
	return f.reader.Read(p)
}

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

func (f *gcsFile) Seek(offset int64, whence int) (int64, error) {
	return 0, fmt.Errorf("seek not supported")
}

func (f *gcsFile) Write(p []byte) (n int, err error) {
	if f.writer == nil {
		return 0, fmt.Errorf("file not opened for writing")
	}
	return f.writer.Write(p)
}

func (f *gcsFile) WriteAt(p []byte, off int64) (n int, err error) {
	return 0, fmt.Errorf("writeat not supported")
}

func (f *gcsFile) Name() string {
	return f.name
}

func (f *gcsFile) Readdir(count int) ([]os.FileInfo, error) {
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
				name:  strings.TrimSuffix(attrs.Prefix, "/"),
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

func (f *gcsFile) Sync() error {
	return nil
}

func (f *gcsFile) Truncate(size int64) error {
	return fmt.Errorf("truncate not supported")
}

func (f *gcsFile) WriteString(s string) (ret int, err error) {
	return f.Write([]byte(s))
}

type gcsFileInfo struct {
	name    string
	size    int64
	modTime time.Time
	isDir   bool
}

func (fi *gcsFileInfo) Name() string {
	return fi.name
}

func (fi *gcsFileInfo) Size() int64 {
	return fi.size
}

func (fi *gcsFileInfo) Mode() os.FileMode {
	if fi.isDir {
		return os.ModeDir | 0755
	}
	return 0644
}

func (fi *gcsFileInfo) ModTime() time.Time {
	return fi.modTime
}

func (fi *gcsFileInfo) IsDir() bool {
	return fi.isDir
}

func (fi *gcsFileInfo) Sys() interface{} {
	return nil
}
