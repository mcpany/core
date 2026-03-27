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
// Summary. Represents a SftpProvider.
type SftpProvider struct {
	fs     afero.Fs
	client *sftp.Client
	conn   *ssh.Client
}

// NewSftpProvider provides newsftpprovider functionality.
//
// Summary: NewSftpProvider.
//
// Parameters.
//   - config: The parameter.
//   - error: The parameter.
//
// Returns.
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

// GetFs provides getfs functionality.
//
// Summary: GetFs.
//
// Parameters.
//   - None.
//
// Returns.
//   - result: The result.
func (p *SftpProvider) GetFs() afero.Fs {
	return p.fs
}

// ResolvePath provides resolvepath functionality.
//
// Summary: ResolvePath.
//
// Parameters.
//   - virtualPath: The parameter.
//   - error: The parameter.
//
// Returns.
//   - None.
func (p *SftpProvider) ResolvePath(virtualPath string) (string, error) {
	// SFTP paths are remote paths. We assume they are absolute or relative to user home.
	// But `clean` is probably good enough for now.
	// NOTE: In the original implementation, SFTP falls through to default in resolvePath, which calls validateLocalPath.
	// THIS WAS LIKELY A BUG as it tried to validate SFTP paths against local root_paths.
	// Here we fix it by just cleaning the path.
	return filepath.Clean(virtualPath), nil
}

// Close provides close functionality.
//
// Summary: Close.
//
// Parameters.
//   - None.
//
// Returns.
//   - result: The result.
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

// Create provides create functionality.
//
// Summary: Create.
//
// Parameters.
//   - name: The parameter.
//   - error: The parameter.
//
// Returns.
//   - None.
func (s *sftpFs) Create(name string) (afero.File, error) {
	f, err := s.client.Create(name)
	if err != nil {
		return nil, err
	}
	return &sftpFile{f: f, client: s.client}, nil
}

// Mkdir provides mkdir functionality.
//
// Summary: Mkdir.
//
// Parameters.
//   - name: The parameter.
//   - _: The parameter.
//
// Returns.
//   - result: The result.
func (s *sftpFs) Mkdir(name string, _ os.FileMode) error {
	return s.client.Mkdir(name)
}

// MkdirAll provides mkdirall functionality.
//
// Summary: MkdirAll.
//
// Parameters.
//   - path: The parameter.
//   - _: The parameter.
//
// Returns.
//   - result: The result.
func (s *sftpFs) MkdirAll(path string, _ os.FileMode) error {
	return s.client.MkdirAll(path)
}

// Open provides open functionality.
//
// Summary: Open.
//
// Parameters.
//   - name: The parameter.
//   - error: The parameter.
//
// Returns.
//   - None.
func (s *sftpFs) Open(name string) (afero.File, error) {
	f, err := s.client.Open(name)
	if err != nil {
		return nil, err
	}
	return &sftpFile{f: f, client: s.client}, nil
}

// OpenFile provides openfile functionality.
//
// Summary: OpenFile.
//
// Parameters.
//   - name: The parameter.
//   - flag: The parameter.
//   - _: The parameter.
//   - error: The parameter.
//
// Returns.
//   - None.
func (s *sftpFs) OpenFile(name string, flag int, _ os.FileMode) (afero.File, error) {
	f, err := s.client.OpenFile(name, flag)
	if err != nil {
		return nil, err
	}
	return &sftpFile{f: f, client: s.client}, nil
}

// Remove provides remove functionality.
//
// Summary: Remove.
//
// Parameters.
//   - name: The parameter.
//
// Returns.
//   - result: The result.
func (s *sftpFs) Remove(name string) error {
	return s.client.Remove(name)
}

// RemoveAll provides removeall functionality.
//
// Summary: RemoveAll.
//
// Parameters.
//   - path: The parameter.
//
// Returns.
//   - result: The result.
func (s *sftpFs) RemoveAll(path string) error {
	// sftp.Client.RemoveAll actually does recursive removal
	return s.client.RemoveAll(path)
}

// Rename provides rename functionality.
//
// Summary: Rename.
//
// Parameters.
//   - oldname: The parameter.
//   - newname: The parameter.
//
// Returns.
//   - result: The result.
func (s *sftpFs) Rename(oldname, newname string) error {
	return s.client.Rename(oldname, newname)
}

// Stat provides stat functionality.
//
// Summary: Stat.
//
// Parameters.
//   - name: The parameter.
//   - error: The parameter.
//
// Returns.
//   - None.
func (s *sftpFs) Stat(name string) (os.FileInfo, error) {
	return s.client.Stat(name)
}

// Name provides name functionality.
//
// Summary: Name.
//
// Parameters.
//   - None.
//
// Returns.
//   - result: The result.
func (s *sftpFs) Name() string {
	return "sftp"
}

// Chmod provides chmod functionality.
//
// Summary: Chmod.
//
// Parameters.
//   - name: The parameter.
//   - mode: The parameter.
//
// Returns.
//   - result: The result.
func (s *sftpFs) Chmod(name string, mode os.FileMode) error {
	return s.client.Chmod(name, mode)
}

// Chown provides chown functionality.
//
// Summary: Chown.
//
// Parameters.
//   - name: The parameter.
//   - uid: The parameter.
//   - gid: The parameter.
//
// Returns.
//   - result: The result.
func (s *sftpFs) Chown(name string, uid, gid int) error {
	return s.client.Chown(name, uid, gid)
}

// Chtimes provides chtimes functionality.
//
// Summary: Chtimes.
//
// Parameters.
//   - name: The parameter.
//   - atime: The parameter.
//   - mtime: The parameter.
//
// Returns.
//   - result: The result.
func (s *sftpFs) Chtimes(name string, atime time.Time, mtime time.Time) error {
	return s.client.Chtimes(name, atime, mtime)
}

type sftpFile struct {
	f      *sftp.File
	client *sftp.Client
}

// Close provides close functionality.
//
// Summary: Close.
//
// Parameters.
//   - None.
//
// Returns.
//   - result: The result.
func (f *sftpFile) Close() error {
	return f.f.Close()
}

// Read provides read functionality.
//
// Summary: Read.
//
// Parameters.
//   - p: The parameter.
//   - err: The parameter.
//
// Returns.
//   - None.
func (f *sftpFile) Read(p []byte) (n int, err error) {
	return f.f.Read(p)
}

// ReadAt provides readat functionality.
//
// Summary: ReadAt.
//
// Parameters.
//   - p: The parameter.
//   - off: The parameter.
//   - err: The parameter.
//
// Returns.
//   - None.
func (f *sftpFile) ReadAt(p []byte, off int64) (n int, err error) {
	return f.f.ReadAt(p, off)
}

// Seek provides seek functionality.
//
// Summary: Seek.
//
// Parameters.
//   - offset: The parameter.
//   - whence: The parameter.
//   - error: The parameter.
//
// Returns.
//   - None.
func (f *sftpFile) Seek(offset int64, whence int) (int64, error) {
	return f.f.Seek(offset, whence)
}

// Write provides write functionality.
//
// Summary: Write.
//
// Parameters.
//   - p: The parameter.
//   - err: The parameter.
//
// Returns.
//   - None.
func (f *sftpFile) Write(p []byte) (n int, err error) {
	return f.f.Write(p)
}

// WriteAt provides writeat functionality.
//
// Summary: WriteAt.
//
// Parameters.
//   - p: The parameter.
//   - off: The parameter.
//   - err: The parameter.
//
// Returns.
//   - None.
func (f *sftpFile) WriteAt(p []byte, off int64) (n int, err error) {
	return f.f.WriteAt(p, off)
}

// Name provides name functionality.
//
// Summary: Name.
//
// Parameters.
//   - None.
//
// Returns.
//   - result: The result.
func (f *sftpFile) Name() string {
	return f.f.Name()
}

// Readdir provides readdir functionality.
//
// Summary: Readdir.
//
// Parameters.
//   - _: The parameter.
//   - error: The parameter.
//
// Returns.
//   - None.
func (f *sftpFile) Readdir(_ int) ([]os.FileInfo, error) {
	return f.client.ReadDir(f.f.Name())
}

// Readdirnames provides readdirnames functionality.
//
// Summary: Readdirnames.
//
// Parameters.
//   - n: The parameter.
//   - error: The parameter.
//
// Returns.
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

// Stat provides stat functionality.
//
// Summary: Stat.
//
// Parameters.
//   - ): The parameter.
//   - error: The parameter.
//
// Returns.
//   - None.
func (f *sftpFile) Stat() (os.FileInfo, error) {
	return f.f.Stat()
}

// Sync provides sync functionality.
//
// Summary: Sync.
//
// Parameters.
//   - None.
//
// Returns.
//   - result: The result.
func (f *sftpFile) Sync() error {
	return nil
}

// Truncate provides truncate functionality.
//
// Summary: Truncate.
//
// Parameters.
//   - size: The parameter.
//
// Returns.
//   - result: The result.
func (f *sftpFile) Truncate(size int64) error {
	return f.f.Truncate(size)
}

// WriteString provides writestring functionality.
//
// Summary: WriteString.
//
// Parameters.
//   - s: The parameter.
//   - err: The parameter.
//
// Returns.
//   - None.
func (f *sftpFile) WriteString(s string) (ret int, err error) {
	return f.f.Write([]byte(s))
}
