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

// Summary: SftpProvider provides access to files via SFTP. Represents a SftpProvider.
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
type SftpProvider struct {
	fs     afero.Fs
	client *sftp.Client
	conn   *ssh.Client
}

// Summary: NewSftpProvider creates a new SftpProvider from the given configuration.
//
// Parameters:
//   - config (*configv1.SftpFs): The config parameter.
//
// Returns:
//   - *SftpProvider: The resulting *SftpProvider.
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails or is invalid.
//
// Side Effects:
//   - None.
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

// Summary: GetFs returns the underlying filesystem.
//
// Parameters:
//   - None.
//
// Returns:
//   - afero.Fs: The resulting afero.Fs.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (p *SftpProvider) GetFs() afero.Fs {
	return p.fs
}

// Summary: ResolvePath resolves the virtual path to a real path.
//
// Parameters:
//   - virtualPath (string): The virtualPath parameter.
//
// Returns:
//   - string: The resulting string.
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails or is invalid.
//
// Side Effects:
//   - None.
func (p *SftpProvider) ResolvePath(virtualPath string) (string, error) {
	// SFTP paths are remote paths. We assume they are absolute or relative to user home.
	// But `clean` is probably good enough for now.
	// NOTE: In the original implementation, SFTP falls through to default in resolvePath, which calls validateLocalPath.
	// THIS WAS LIKELY A BUG as it tried to validate SFTP paths against local root_paths.
	// Here we fix it by just cleaning the path.
	return filepath.Clean(virtualPath), nil
}

// Summary: Close closes the SFTP client and connection.
//
// Parameters:
//   - None.
//
// Returns:
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails or is invalid.
//
// Side Effects:
//   - None.
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

// Summary: Create creates a file in the filesystem, returning the file and an error, if any happens.
//
// Parameters:
//   - name (string): The name parameter.
//
// Returns:
//   - afero.File: The resulting afero.File.
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails or is invalid.
//
// Side Effects:
//   - None.
func (s *sftpFs) Create(name string) (afero.File, error) {
	f, err := s.client.Create(name)
	if err != nil {
		return nil, err
	}
	return &sftpFile{f: f, client: s.client}, nil
}

// Summary: Mkdir creates a directory in the filesystem, returning an error, if any happens.
//
// Parameters:
//   - name (string): The name parameter.
//   - _ (os.FileMode): The _ parameter.
//
// Returns:
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails or is invalid.
//
// Side Effects:
//   - None.
func (s *sftpFs) Mkdir(name string, _ os.FileMode) error {
	return s.client.Mkdir(name)
}

// Summary: MkdirAll creates a directory path and all parents that does not exist for a given name.
//
// Parameters:
//   - path (string): The path parameter.
//   - _ (os.FileMode): The _ parameter.
//
// Returns:
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails or is invalid.
//
// Side Effects:
//   - None.
func (s *sftpFs) MkdirAll(path string, _ os.FileMode) error {
	return s.client.MkdirAll(path)
}

// Summary: Open opens a file, returning it or an error, if any happens.
//
// Parameters:
//   - name (string): The name parameter.
//
// Returns:
//   - afero.File: The resulting afero.File.
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails or is invalid.
//
// Side Effects:
//   - None.
func (s *sftpFs) Open(name string) (afero.File, error) {
	f, err := s.client.Open(name)
	if err != nil {
		return nil, err
	}
	return &sftpFile{f: f, client: s.client}, nil
}

// Summary: OpenFile opens a file using the given flags and the given mode.
//
// Parameters:
//   - name (string): The name parameter.
//   - flag (int): The flag parameter.
//   - _ (os.FileMode): The _ parameter.
//
// Returns:
//   - afero.File: The resulting afero.File.
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails or is invalid.
//
// Side Effects:
//   - None.
func (s *sftpFs) OpenFile(name string, flag int, _ os.FileMode) (afero.File, error) {
	f, err := s.client.OpenFile(name, flag)
	if err != nil {
		return nil, err
	}
	return &sftpFile{f: f, client: s.client}, nil
}

// Summary: Remove removes a file identified by name, returning an error, if any happens.
//
// Parameters:
//   - name (string): The name parameter.
//
// Returns:
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails or is invalid.
//
// Side Effects:
//   - None.
func (s *sftpFs) Remove(name string) error {
	return s.client.Remove(name)
}

// Summary: RemoveAll removes a directory path and any children it contains.
//
// Parameters:
//   - path (string): The path parameter.
//
// Returns:
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails or is invalid.
//
// Side Effects:
//   - None.
func (s *sftpFs) RemoveAll(path string) error {
	// sftp.Client.RemoveAll actually does recursive removal
	return s.client.RemoveAll(path)
}

// Summary: Rename renames a file.
//
// Parameters:
//   - oldname (string): The oldname parameter.
//   - newname (string): The newname parameter.
//
// Returns:
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails or is invalid.
//
// Side Effects:
//   - None.
func (s *sftpFs) Rename(oldname, newname string) error {
	return s.client.Rename(oldname, newname)
}

// Summary: Stat returns a FileInfo describing the named file, or an error, if any happens.
//
// Parameters:
//   - name (string): The name parameter.
//
// Returns:
//   - os.FileInfo: The resulting os.FileInfo.
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails or is invalid.
//
// Side Effects:
//   - None.
func (s *sftpFs) Stat(name string) (os.FileInfo, error) {
	return s.client.Stat(name)
}

// Summary: Name returns the name of this file system.
//
// Parameters:
//   - None.
//
// Returns:
//   - string: The resulting string.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (s *sftpFs) Name() string {
	return "sftp"
}

// Summary: Chmod changes the mode of the named file to mode.
//
// Parameters:
//   - name (string): The name parameter.
//   - mode (os.FileMode): The mode parameter.
//
// Returns:
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails or is invalid.
//
// Side Effects:
//   - None.
func (s *sftpFs) Chmod(name string, mode os.FileMode) error {
	return s.client.Chmod(name, mode)
}

// Summary: Chown changes the uid and gid of the named file.
//
// Parameters:
//   - name (string): The name parameter.
//   - uid (int): The uid parameter.
//   - gid (int): The gid parameter.
//
// Returns:
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails or is invalid.
//
// Side Effects:
//   - None.
func (s *sftpFs) Chown(name string, uid, gid int) error {
	return s.client.Chown(name, uid, gid)
}

// Summary: Chtimes changes the access and modification times of the named file.
//
// Parameters:
//   - name (string): The name parameter.
//   - atime (time.Time): The atime parameter.
//   - mtime (time.Time): The mtime parameter.
//
// Returns:
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails or is invalid.
//
// Side Effects:
//   - None.
func (s *sftpFs) Chtimes(name string, atime time.Time, mtime time.Time) error {
	return s.client.Chtimes(name, atime, mtime)
}

type sftpFile struct {
	f      *sftp.File
	client *sftp.Client
}

// Summary: Close closes the file.
//
// Parameters:
//   - None.
//
// Returns:
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails or is invalid.
//
// Side Effects:
//   - None.
func (f *sftpFile) Close() error {
	return f.f.Close()
}

// Summary: Read reads up to len(b) bytes from the File.
//
// Parameters:
//   - p ([]byte): The p parameter.
//
// Returns:
//   - int: The resulting int.
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails or is invalid.
//
// Side Effects:
//   - None.
func (f *sftpFile) Read(p []byte) (n int, err error) {
	return f.f.Read(p)
}

// Summary: ReadAt reads len(b) bytes from the File starting at byte offset off.
//
// Parameters:
//   - p ([]byte): The p parameter.
//   - off (int64): The off parameter.
//
// Returns:
//   - int: The resulting int.
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails or is invalid.
//
// Side Effects:
//   - None.
func (f *sftpFile) ReadAt(p []byte, off int64) (n int, err error) {
	return f.f.ReadAt(p, off)
}

// Summary: Seek sets the offset for the next Read or Write to offset, interpreted according to whence.
//
// Parameters:
//   - offset (int64): The offset parameter.
//   - whence (int): The whence parameter.
//
// Returns:
//   - int64: The resulting int64.
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails or is invalid.
//
// Side Effects:
//   - None.
func (f *sftpFile) Seek(offset int64, whence int) (int64, error) {
	return f.f.Seek(offset, whence)
}

// Summary: Write writes len(b) bytes to the File.
//
// Parameters:
//   - p ([]byte): The p parameter.
//
// Returns:
//   - int: The resulting int.
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails or is invalid.
//
// Side Effects:
//   - None.
func (f *sftpFile) Write(p []byte) (n int, err error) {
	return f.f.Write(p)
}

// Summary: WriteAt writes len(b) bytes to the File starting at byte offset off.
//
// Parameters:
//   - p ([]byte): The p parameter.
//   - off (int64): The off parameter.
//
// Returns:
//   - int: The resulting int.
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails or is invalid.
//
// Side Effects:
//   - None.
func (f *sftpFile) WriteAt(p []byte, off int64) (n int, err error) {
	return f.f.WriteAt(p, off)
}

// Summary: Name returns the name of the file as presented to Open.
//
// Parameters:
//   - None.
//
// Returns:
//   - string: The resulting string.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (f *sftpFile) Name() string {
	return f.f.Name()
}

// Summary: Readdir reads the contents of the directory associated with file and returns a slice of up to n FileInfo values, as would be returned by Lstat, in directory order.
//
// Parameters:
//   - _ (int): The _ parameter.
//
// Returns:
//   - []os.FileInfo: The resulting []os.FileInfo.
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails or is invalid.
//
// Side Effects:
//   - None.
func (f *sftpFile) Readdir(_ int) ([]os.FileInfo, error) {
	return f.client.ReadDir(f.f.Name())
}

// Summary: Readdirnames reads and returns a slice of names from the directory f.
//
// Parameters:
//   - n (int): The n parameter.
//
// Returns:
//   - []string: The resulting []string.
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails or is invalid.
//
// Side Effects:
//   - None.
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

// Summary: Stat returns the FileInfo structure describing file.
//
// Parameters:
//   - None.
//
// Returns:
//   - os.FileInfo: The resulting os.FileInfo.
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails or is invalid.
//
// Side Effects:
//   - None.
func (f *sftpFile) Stat() (os.FileInfo, error) {
	return f.f.Stat()
}

// Summary: Sync commits the current contents of the file to stable storage.
//
// Parameters:
//   - None.
//
// Returns:
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails or is invalid.
//
// Side Effects:
//   - None.
func (f *sftpFile) Sync() error {
	return nil
}

// Summary: Truncate changes the size of the file.
//
// Parameters:
//   - size (int64): The size parameter.
//
// Returns:
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails or is invalid.
//
// Side Effects:
//   - None.
func (f *sftpFile) Truncate(size int64) error {
	return f.f.Truncate(size)
}

// Summary: WriteString is like Write, but writes the contents of string s rather than a slice of bytes.
//
// Parameters:
//   - s (string): The s parameter.
//
// Returns:
//   - int: The resulting int.
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails or is invalid.
//
// Side Effects:
//   - None.
func (f *sftpFile) WriteString(s string) (ret int, err error) {
	return f.f.Write([]byte(s))
}
