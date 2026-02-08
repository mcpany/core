// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/pkg/sftp"
	"github.com/spf13/afero"
	"golang.org/x/crypto/ssh"
)

// SftpProvider provides access to files via SFTP.
//
// Summary: provides access to files via SFTP.
type SftpProvider struct {
	fs     afero.Fs
	client *sftp.Client
	conn   *ssh.Client
}

// NewSftpProvider creates a new SftpProvider from the given configuration.
//
// Summary: creates a new SftpProvider from the given configuration.
//
// Parameters:
//   - config: *configv1.SftpFs. The config.
//
// Returns:
//   - *SftpProvider: The *SftpProvider.
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
func NewSftpProvider(config *configv1.SftpFs) (*SftpProvider, error) {
	if config == nil {
		return nil, fmt.Errorf("sftp config is nil")
	}

	auths := []ssh.AuthMethod{}
	if config.GetPassword() != "" {
		auths = append(auths, ssh.Password(config.GetPassword()))
	}
	if config.GetKeyPath() != "" {
		key, err := os.ReadFile(config.GetKeyPath())
		if err != nil {
			return nil, fmt.Errorf("failed to read private key: %w", err)
		}
		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}
		auths = append(auths, ssh.PublicKeys(signer))
	}

	sshConfig := &ssh.ClientConfig{
		User: config.GetUsername(),
		Auth: auths,
		//nolint:gosec // user configuration allows connection to arbitrary hosts
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}

	addr := config.GetAddress()
	if !strings.Contains(addr, ":") {
		addr += ":22"
	}

	conn, err := ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to dial ssh: %w", err)
	}

	client, err := sftp.NewClient(conn)
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("failed to create sftp client: %w", err)
	}

	return &SftpProvider{
		fs:     &sftpFs{client: client},
		client: client,
		conn:   conn,
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
func (p *SftpProvider) GetFs() afero.Fs {
	return p.fs
}

// ResolvePath resolves the virtual path to a real path.
//
// Summary: resolves the virtual path to a real path.
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
func (p *SftpProvider) ResolvePath(virtualPath string) (string, error) {
	// SFTP paths are remote paths. We assume they are absolute or relative to user home.
	// But `clean` is probably good enough for now.
	// NOTE: In the original implementation, SFTP falls through to default in resolvePath, which calls validateLocalPath.
	// THIS WAS LIKELY A BUG as it tried to validate SFTP paths against local root_paths.
	// Here we fix it by just cleaning the path.
	return filepath.Clean(virtualPath), nil
}

// Close closes the SFTP client and connection.
//
// Summary: closes the SFTP client and connection.
//
// Parameters:
//   None.
//
// Returns:
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
func (p *SftpProvider) Close() error {
	if p.client != nil {
		_ = p.client.Close()
	}
	if p.conn != nil {
		_ = p.conn.Close()
	}
	return nil
}

// sftpFs implementation copy from original sftp.go

type sftpFs struct {
	client *sftp.Client
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
func (s *sftpFs) Create(name string) (afero.File, error) {
	f, err := s.client.Create(name)
	if err != nil {
		return nil, err
	}
	return &sftpFile{f: f, client: s.client}, nil
}

// Mkdir creates a directory in the filesystem, returning an error, if any happens.
//
// Summary: creates a directory in the filesystem, returning an error, if any happens.
//
// Parameters:
//   - name: string. The name.
//   - _: os.FileMode. The _.
//
// Returns:
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
func (s *sftpFs) Mkdir(name string, _ os.FileMode) error {
	return s.client.Mkdir(name)
}

// MkdirAll creates a directory path and all parents that does not exist for a given name.
//
// Summary: creates a directory path and all parents that does not exist for a given name.
//
// Parameters:
//   - path: string. The path.
//   - _: os.FileMode. The _.
//
// Returns:
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
func (s *sftpFs) MkdirAll(path string, _ os.FileMode) error {
	return s.client.MkdirAll(path)
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
func (s *sftpFs) Open(name string) (afero.File, error) {
	f, err := s.client.Open(name)
	if err != nil {
		return nil, err
	}
	return &sftpFile{f: f, client: s.client}, nil
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
func (s *sftpFs) OpenFile(name string, flag int, _ os.FileMode) (afero.File, error) {
	f, err := s.client.OpenFile(name, flag)
	if err != nil {
		return nil, err
	}
	return &sftpFile{f: f, client: s.client}, nil
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
func (s *sftpFs) Remove(name string) error {
	return s.client.Remove(name)
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
func (s *sftpFs) RemoveAll(path string) error {
	// sftp.Client.RemoveAll actually does recursive removal
	return s.client.RemoveAll(path)
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
func (s *sftpFs) Rename(oldname, newname string) error {
	return s.client.Rename(oldname, newname)
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
func (s *sftpFs) Stat(name string) (os.FileInfo, error) {
	return s.client.Stat(name)
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
func (s *sftpFs) Name() string {
	return "sftp"
}

// Chmod changes the mode of the named file to mode.
//
// Summary: changes the mode of the named file to mode.
//
// Parameters:
//   - name: string. The name.
//   - mode: os.FileMode. The mode.
//
// Returns:
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
func (s *sftpFs) Chmod(name string, mode os.FileMode) error {
	return s.client.Chmod(name, mode)
}

// Chown changes the uid and gid of the named file.
//
// Summary: changes the uid and gid of the named file.
//
// Parameters:
//   - name: string. The name.
//   - uid: int. The uid.
//   - gid: int. The gid.
//
// Returns:
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
func (s *sftpFs) Chown(name string, uid, gid int) error {
	return s.client.Chown(name, uid, gid)
}

// Chtimes changes the access and modification times of the named file.
//
// Summary: changes the access and modification times of the named file.
//
// Parameters:
//   - name: string. The name.
//   - atime: time.Time. The atime.
//   - mtime: time.Time. The mtime.
//
// Returns:
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
func (s *sftpFs) Chtimes(name string, atime time.Time, mtime time.Time) error {
	return s.client.Chtimes(name, atime, mtime)
}

type sftpFile struct {
	f      *sftp.File
	client *sftp.Client
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
func (f *sftpFile) Close() error {
	return f.f.Close()
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
func (f *sftpFile) Read(p []byte) (n int, err error) {
	return f.f.Read(p)
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
func (f *sftpFile) ReadAt(p []byte, off int64) (n int, err error) {
	return f.f.ReadAt(p, off)
}

// Seek sets the offset for the next Read or Write to offset, interpreted according to whence.
//
// Summary: sets the offset for the next Read or Write to offset, interpreted according to whence.
//
// Parameters:
//   - offset: int64. The offset.
//   - whence: int. The whence.
//
// Returns:
//   - int64: The int64.
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
func (f *sftpFile) Seek(offset int64, whence int) (int64, error) {
	return f.f.Seek(offset, whence)
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
func (f *sftpFile) Write(p []byte) (n int, err error) {
	return f.f.Write(p)
}

// WriteAt writes len(b) bytes to the File starting at byte offset off.
//
// Summary: writes len(b) bytes to the File starting at byte offset off.
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
func (f *sftpFile) WriteAt(p []byte, off int64) (n int, err error) {
	return f.f.WriteAt(p, off)
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
func (f *sftpFile) Name() string {
	return f.f.Name()
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
func (f *sftpFile) Readdir(_ int) ([]os.FileInfo, error) {
	return f.client.ReadDir(f.f.Name())
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
func (f *sftpFile) Readdirnames(n int) ([]string, error) {
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
func (f *sftpFile) Stat() (os.FileInfo, error) {
	return f.f.Stat()
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
func (f *sftpFile) Sync() error {
	return nil
}

// Truncate changes the size of the file.
//
// Summary: changes the size of the file.
//
// Parameters:
//   - size: int64. The size.
//
// Returns:
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
func (f *sftpFile) Truncate(size int64) error {
	return f.f.Truncate(size)
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
func (f *sftpFile) WriteString(s string) (ret int, err error) {
	return f.f.Write([]byte(s))
}
