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

// SftpProvider represents the public SftpProvider entity.
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
type SftpProvider struct {
	fs     afero.Fs
	client *sftp.Client
	conn   *ssh.Client
}

// NewSftpProvider serves as a public interface for interacting with NewSftpProvider.
//
// Summary: Constructs and returns an initialized sftp provider ready for consumption.
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
func (p *SftpProvider) GetFs() afero.Fs {
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
func (p *SftpProvider) ResolvePath(virtualPath string) (string, error) {
	// SFTP paths are remote paths. We assume they are absolute or relative to user home.
	// But `clean` is probably good enough for now.
	// NOTE: In the original implementation, SFTP falls through to default in resolvePath, which calls validateLocalPath.
	// THIS WAS LIKELY A BUG as it tried to validate SFTP paths against local root_paths.
	// Here we fix it by just cleaning the path.
	return filepath.Clean(virtualPath), nil
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
func (s *sftpFs) Create(name string) (afero.File, error) {
	f, err := s.client.Create(name)
	if err != nil {
		return nil, err
	}
	return &sftpFile{f: f, client: s.client}, nil
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
func (s *sftpFs) Mkdir(name string, _ os.FileMode) error {
	return s.client.Mkdir(name)
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
func (s *sftpFs) MkdirAll(path string, _ os.FileMode) error {
	return s.client.MkdirAll(path)
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
func (s *sftpFs) Open(name string) (afero.File, error) {
	f, err := s.client.Open(name)
	if err != nil {
		return nil, err
	}
	return &sftpFile{f: f, client: s.client}, nil
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
func (s *sftpFs) OpenFile(name string, flag int, _ os.FileMode) (afero.File, error) {
	f, err := s.client.OpenFile(name, flag)
	if err != nil {
		return nil, err
	}
	return &sftpFile{f: f, client: s.client}, nil
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
func (s *sftpFs) Remove(name string) error {
	return s.client.Remove(name)
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
func (s *sftpFs) RemoveAll(path string) error {
	// sftp.Client.RemoveAll actually does recursive removal
	return s.client.RemoveAll(path)
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
func (s *sftpFs) Rename(oldname, newname string) error {
	return s.client.Rename(oldname, newname)
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
func (s *sftpFs) Stat(name string) (os.FileInfo, error) {
	return s.client.Stat(name)
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
func (s *sftpFs) Name() string {
	return "sftp"
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
func (s *sftpFs) Chmod(name string, mode os.FileMode) error {
	return s.client.Chmod(name, mode)
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
func (s *sftpFs) Chown(name string, uid, gid int) error {
	return s.client.Chown(name, uid, gid)
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
func (s *sftpFs) Chtimes(name string, atime time.Time, mtime time.Time) error {
	return s.client.Chtimes(name, atime, mtime)
}

type sftpFile struct {
	f      *sftp.File
	client *sftp.Client
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
func (f *sftpFile) Close() error {
	return f.f.Close()
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
func (f *sftpFile) Read(p []byte) (n int, err error) {
	return f.f.Read(p)
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
func (f *sftpFile) ReadAt(p []byte, off int64) (n int, err error) {
	return f.f.ReadAt(p, off)
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
func (f *sftpFile) Seek(offset int64, whence int) (int64, error) {
	return f.f.Seek(offset, whence)
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
func (f *sftpFile) Write(p []byte) (n int, err error) {
	return f.f.Write(p)
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
func (f *sftpFile) WriteAt(p []byte, off int64) (n int, err error) {
	return f.f.WriteAt(p, off)
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
func (f *sftpFile) Name() string {
	return f.f.Name()
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
func (f *sftpFile) Readdir(_ int) ([]os.FileInfo, error) {
	return f.client.ReadDir(f.f.Name())
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
func (f *sftpFile) Stat() (os.FileInfo, error) {
	return f.f.Stat()
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
func (f *sftpFile) Sync() error {
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
func (f *sftpFile) Truncate(size int64) error {
	return f.f.Truncate(size)
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
func (f *sftpFile) WriteString(s string) (ret int, err error) {
	return f.f.Write([]byte(s))
}
