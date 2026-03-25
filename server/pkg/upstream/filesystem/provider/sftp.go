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
// Summary: Represents a SftpProvider.
type SftpProvider struct {
	fs     afero.Fs
	client *sftp.Client
	conn   *ssh.Client
}

// NewSftpProvider creates a new sftp provider.
//
// Summary: Creates a new sftp provider.
//
// Parameters: - None.
//   - config (*configv1.SftpFs): The config.
//
// Returns: - None.
//   - *SftpProvider: The result.
//   - error: An error if the operation fails.
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

// GetFs retrieves the fs.
//
// Summary: Retrieves the fs.
//
// Parameters: - None.
//   - None.
//
// Returns: - None.
//   - afero.Fs: The result.
func (p *SftpProvider) GetFs() afero.Fs {
	return p.fs
}

// ResolvePath resolvePath resolve path.
//
// Summary: ResolvePath resolve path.
//
// Parameters: - None.
//   - virtualPath (string): The virtual path.
//
// Returns: - None.
//   - string: The result.
//   - error: An error if the operation fails.
func (p *SftpProvider) ResolvePath(virtualPath string) (string, error) {
	// SFTP paths are remote paths. We assume they are absolute or relative to user home.
	// But `clean` is probably good enough for now.
	// NOTE: In the original implementation, SFTP falls through to default in resolvePath, which calls validateLocalPath.
	// THIS WAS LIKELY A BUG as it tried to validate SFTP paths against local root_paths.
	// Here we fix it by just cleaning the path.
	return filepath.Clean(virtualPath), nil
}

// Close close close.
//
// Summary: Close close.
//
// Parameters: - None.
//   - None.
//
// Returns: - None.
//   - error: An error if the operation fails.
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

// Create persists the .
//
// Summary: Persists the .
//
// Parameters: - None.
//   - name (string): The name.
//
// Returns: - None.
//   - afero.File: The result.
//   - error: An error if the operation fails.
func (s *sftpFs) Create(name string) (afero.File, error) {
	f, err := s.client.Create(name)
	if err != nil {
		return nil, err
	}
	return &sftpFile{f: f, client: s.client}, nil
}

// Mkdir mkdir mkdir.
//
// Summary: Mkdir mkdir.
//
// Parameters: - None.
//   - name (string): The name.
//   - _ (os.FileMode): Unused parameter.
//
// Returns: - None.
//   - error: An error if the operation fails.
func (s *sftpFs) Mkdir(name string, _ os.FileMode) error {
	return s.client.Mkdir(name)
}

// MkdirAll mkdirAll mkdir all.
//
// Summary: MkdirAll mkdir all.
//
// Parameters: - None.
//   - path (string): The path.
//   - _ (os.FileMode): Unused parameter.
//
// Returns: - None.
//   - error: An error if the operation fails.
func (s *sftpFs) MkdirAll(path string, _ os.FileMode) error {
	return s.client.MkdirAll(path)
}

// Open open open.
//
// Summary: Open open.
//
// Parameters: - None.
//   - name (string): The name.
//
// Returns: - None.
//   - afero.File: The result.
//   - error: An error if the operation fails.
func (s *sftpFs) Open(name string) (afero.File, error) {
	f, err := s.client.Open(name)
	if err != nil {
		return nil, err
	}
	return &sftpFile{f: f, client: s.client}, nil
}

// OpenFile openFile open file.
//
// Summary: OpenFile open file.
//
// Parameters: - None.
//   - name (string): The name.
//   - flag (int): The flag.
//   - _ (os.FileMode): Unused parameter.
//
// Returns: - None.
//   - afero.File: The result.
//   - error: An error if the operation fails.
func (s *sftpFs) OpenFile(name string, flag int, _ os.FileMode) (afero.File, error) {
	f, err := s.client.OpenFile(name, flag)
	if err != nil {
		return nil, err
	}
	return &sftpFile{f: f, client: s.client}, nil
}

// Remove remove remove.
//
// Summary: Remove remove.
//
// Parameters: - None.
//   - name (string): The name.
//
// Returns: - None.
//   - error: An error if the operation fails.
func (s *sftpFs) Remove(name string) error {
	return s.client.Remove(name)
}

// RemoveAll removeAll remove all.
//
// Summary: RemoveAll remove all.
//
// Parameters: - None.
//   - path (string): The path.
//
// Returns: - None.
//   - error: An error if the operation fails.
func (s *sftpFs) RemoveAll(path string) error {
	// sftp.Client.RemoveAll actually does recursive removal
	return s.client.RemoveAll(path)
}

// Rename rename rename.
//
// Summary: Rename rename.
//
// Parameters: - None.
//   - oldname (unknown): The oldname.
//   - newname (string): The newname.
//
// Returns: - None.
//   - error: An error if the operation fails.
func (s *sftpFs) Rename(oldname, newname string) error {
	return s.client.Rename(oldname, newname)
}

// Stat stat stat.
//
// Summary: Stat stat.
//
// Parameters: - None.
//   - name (string): The name.
//
// Returns: - None.
//   - os.FileInfo: The result.
//   - error: An error if the operation fails.
func (s *sftpFs) Stat(name string) (os.FileInfo, error) {
	return s.client.Stat(name)
}

// Name name name.
//
// Summary: Name name.
//
// Parameters: - None.
//   - None.
//
// Returns: - None.
//   - string: The result.
func (s *sftpFs) Name() string {
	return "sftp"
}

// Chmod chmod chmod.
//
// Summary: Chmod chmod.
//
// Parameters: - None.
//   - name (string): The name.
//   - mode (os.FileMode): The mode.
//
// Returns: - None.
//   - error: An error if the operation fails.
func (s *sftpFs) Chmod(name string, mode os.FileMode) error {
	return s.client.Chmod(name, mode)
}

// Chown chown chown.
//
// Summary: Chown chown.
//
// Parameters: - None.
//   - name (string): The name.
//   - uid (unknown): The uid.
//   - gid (int): The gid.
//
// Returns: - None.
//   - error: An error if the operation fails.
func (s *sftpFs) Chown(name string, uid, gid int) error {
	return s.client.Chown(name, uid, gid)
}

// Chtimes chtimes chtimes.
//
// Summary: Chtimes chtimes.
//
// Parameters: - None.
//   - name (string): The name.
//   - atime (time.Time): The atime.
//   - mtime (time.Time): The mtime.
//
// Returns: - None.
//   - error: An error if the operation fails.
func (s *sftpFs) Chtimes(name string, atime time.Time, mtime time.Time) error {
	return s.client.Chtimes(name, atime, mtime)
}

type sftpFile struct {
	f      *sftp.File
	client *sftp.Client
}

// Close close close.
//
// Summary: Close close.
//
// Parameters: - None.
//   - None.
//
// Returns: - None.
//   - error: An error if the operation fails.
func (f *sftpFile) Close() error {
	return f.f.Close()
}

// Read read read.
//
// Summary: Read read.
//
// Parameters: - None.
//   - p ([]byte): The p.
//
// Returns: - None.
//   - int: The result.
//   - error: An error if the operation fails.
func (f *sftpFile) Read(p []byte) (n int, err error) {
	return f.f.Read(p)
}

// ReadAt readAt read at.
//
// Summary: ReadAt read at.
//
// Parameters: - None.
//   - p ([]byte): The p.
//   - off (int64): The off.
//
// Returns: - None.
//   - int: The result.
//   - error: An error if the operation fails.
func (f *sftpFile) ReadAt(p []byte, off int64) (n int, err error) {
	return f.f.ReadAt(p, off)
}

// Seek seek seek.
//
// Summary: Seek seek.
//
// Parameters: - None.
//   - offset (int64): The offset.
//   - whence (int): The whence.
//
// Returns: - None.
//   - int64: The result.
//   - error: An error if the operation fails.
func (f *sftpFile) Seek(offset int64, whence int) (int64, error) {
	return f.f.Seek(offset, whence)
}

// Write write write.
//
// Summary: Write write.
//
// Parameters: - None.
//   - p ([]byte): The p.
//
// Returns: - None.
//   - int: The result.
//   - error: An error if the operation fails.
func (f *sftpFile) Write(p []byte) (n int, err error) {
	return f.f.Write(p)
}

// WriteAt writeAt write at.
//
// Summary: WriteAt write at.
//
// Parameters: - None.
//   - p ([]byte): The p.
//   - off (int64): The off.
//
// Returns: - None.
//   - int: The result.
//   - error: An error if the operation fails.
func (f *sftpFile) WriteAt(p []byte, off int64) (n int, err error) {
	return f.f.WriteAt(p, off)
}

// Name name name.
//
// Summary: Name name.
//
// Parameters: - None.
//   - None.
//
// Returns: - None.
//   - string: The result.
func (f *sftpFile) Name() string {
	return f.f.Name()
}

// Readdir readdir readdir.
//
// Summary: Readdir readdir.
//
// Parameters: - None.
//   - _ (int): Unused parameter.
//
// Returns: - None.
//   - []os.FileInfo: The result.
//   - error: An error if the operation fails.
func (f *sftpFile) Readdir(_ int) ([]os.FileInfo, error) {
	return f.client.ReadDir(f.f.Name())
}

// Readdirnames readdirnames readdirnames.
//
// Summary: Readdirnames readdirnames.
//
// Parameters: - None.
//   - n (int): The n.
//
// Returns: - None.
//   - []string: The result.
//   - error: An error if the operation fails.
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

// Stat stat stat.
//
// Summary: Stat stat.
//
// Parameters: - None.
//   - None.
//
// Returns: - None.
//   - os.FileInfo: The result.
//   - error: An error if the operation fails.
func (f *sftpFile) Stat() (os.FileInfo, error) {
	return f.f.Stat()
}

// Sync sync sync.
//
// Summary: Sync sync.
//
// Parameters: - None.
//   - None.
//
// Returns: - None.
//   - error: An error if the operation fails.
func (f *sftpFile) Sync() error {
	return nil
}

// Truncate truncate truncate.
//
// Summary: Truncate truncate.
//
// Parameters: - None.
//   - size (int64): The size.
//
// Returns: - None.
//   - error: An error if the operation fails.
func (f *sftpFile) Truncate(size int64) error {
	return f.f.Truncate(size)
}

// WriteString writeString write string.
//
// Summary: WriteString write string.
//
// Parameters: - None.
//   - s (string): The s.
//
// Returns: - None.
//   - int: The result.
//   - error: An error if the operation fails.
func (f *sftpFile) WriteString(s string) (ret int, err error) {
	return f.f.Write([]byte(s))
}
