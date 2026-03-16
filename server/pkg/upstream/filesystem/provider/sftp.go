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
// Summary: SftpProvider provides access to files via SFTP.
//
// Summary: SftpProvider provides access to files via SFTP.
type SftpProvider struct {
	fs     afero.Fs
	client *sftp.Client
	conn   *ssh.Client
// NewSftpProvider creates a new SftpProvider from the given configuration.
//
// Summary: NewSftpProvider creates a new SftpProvider from the given configuration.
//
// Parameters:
//   - config (*configv1.SftpFs): The configuration settings.
//
// Returns:
//   - *SftpProvider: The resulting object or data structure.
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
// ResolvePath resolves the virtual path to a real path.
//
// Summary: ResolvePath resolves the virtual path to a real path.
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
// Summary: ResolvePath resolves the virtual path to a real path.
//
// Parameters:
//   - virtualPath (string): The textual representation of virtualpath.
//
// Returns:
//   - string: The resulting text.
//   - error: An error if the execution fails, otherwise nil.
//
// Close closes the SFTP client and connection.
//
// Summary: Close closes the SFTP client and connection.
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
	// Here we fix it by just cleaning the path.
	return filepath.Clean(virtualPath), nil
}

// Close closes the SFTP client and connection.
//
// Summary: Close closes the SFTP client and connection.
//
// Parameters:
//   - None.
//
// Returns:
//   - error: An error if the execution fails, otherwise nil.
//
// Errors:
//   - Returns an error if the operation fails, invalid input is provided, or a downstream dependency fails.
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

type sftpFs struct {
	client *sftp.Client
}

// Create creates a file in the filesystem, returning the file and an error, if any happens.
//
// Summary: Create creates a file in the filesystem, returning the file and an error, if any happens.
// Mkdir creates a directory in the filesystem, returning an error, if any happens.
//
// Summary: Mkdir creates a directory in the filesystem, returning an error, if any happens.
//
// Parameters:
//   - name (string): The human-readable or system name.
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
	f, err := s.client.Create(name)
	if err != nil {
		return nil, err
	}
// MkdirAll creates a directory path and all parents that does not exist for a given name.
//
// Summary: MkdirAll creates a directory path and all parents that does not exist for a given name.
//
// Parameters:
//   - path (string): The textual representation of path.
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
//
// Side Effects:
//   - May modify internal state or perform external network calls.
func (s *sftpFs) MkdirAll(path string, _ os.FileMode) error {
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
//
// Side Effects:
//   - May modify internal state or perform external network calls.
func (s *sftpFs) Open(name string) (afero.File, error) {
	f, err := s.client.Open(name)
	if err != nil {
		return nil, err
	}
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
//   - afero.File: The resulting object or data structure.
//   - error: An error if the execution fails, otherwise nil.
//
// Errors:
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
//
// Summary: RemoveAll removes a directory path and any children it contains.
//
// Parameters:
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

// Rename renames a file.
//
// Summary: Rename renames a file.
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
//
// Errors:
//   - Returns an error if the operation fails, invalid input is provided, or a downstream dependency fails.
//
// Chmod changes the mode of the named file to mode.
//
// Summary: Chmod changes the mode of the named file to mode.
//
// Parameters:
//   - name (string): The human-readable or system name.
//   - mode (os.FileMode): The provided mode data.
//
// Returns:
//   - error: An error if the execution fails, otherwise nil.
//
// Errors:
//   - Returns an error if the operation fails, invalid input is provided, or a downstream dependency fails.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
//   - os.FileInfo: The resulting object or data structure.
//   - error: An error if the execution fails, otherwise nil.
//
// Errors:
// Chown changes the uid and gid of the named file.
//
// Summary: Chown changes the uid and gid of the named file.
//
// Parameters:
//   - name (string): The human-readable or system name.
//   - uid (int): The numeric value for uid.
//   - gid (int): The numeric value for gid.
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
//   - name (string): The human-readable or system name.
//   - atime (time.Time): The provided atime data.
//   - mtime (time.Time): The provided mtime data.
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
//   - uid (int): The numeric value for uid.
//   - gid (int): The numeric value for gid.
//
// Returns:
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
// Parameters:
//   - name (string): The human-readable or system name.
//   - atime (time.Time): The provided atime data.
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
	f      *sftp.File
	client *sftp.Client
}

// Seek sets the offset for the next Read or Write to offset, interpreted according to whence.
//
// Summary: Seek sets the offset for the next Read or Write to offset, interpreted according to whence.
//
// Parameters:
//   - offset (int64): The numeric value for offset.
//   - whence (int): The numeric value for whence.
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
func (f *sftpFile) Close() error {
	return f.f.Close()
}

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
// Side Effects:
//   - May modify internal state or perform external network calls.
func (f *sftpFile) Read(p []byte) (n int, err error) {
	return f.f.Read(p)
// WriteAt writes len(b) bytes to the File starting at byte offset off.
//
// Summary: WriteAt writes len(b) bytes to the File starting at byte offset off.
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
//   - Returns an error if the operation fails, invalid input is provided, or a downstream dependency fails.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
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
//
// Parameters:
//   - offset (int64): The numeric value for offset.
//   - whence (int): The numeric value for whence.
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
//
// Summary: Write writes len(b) bytes to the File.
//
// Parameters:
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

// WriteAt writes len(b) bytes to the File starting at byte offset off.
//
// Summary: WriteAt writes len(b) bytes to the File starting at byte offset off.
//
// Parameters:
//   - p ([]byte): The provided p data.
//   - off (int64): The numeric value for off.
//
// Returns:
//   - n (int): The calculated numeric value.
//   - err (error): An error if the execution fails, otherwise nil.
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
//
// Summary: Name returns the name of the file as presented to Open.
//
// Parameters:
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
func (f *sftpFile) Name() string {
	return f.f.Name()
}

// Truncate changes the size of the file.
//
// Summary: Truncate changes the size of the file.
//
// Parameters:
//   - size (int64): The numeric value for size.
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
func (f *sftpFile) Readdir(_ int) ([]os.FileInfo, error) {
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
// Errors:
//   - Returns an error if the operation fails, invalid input is provided, or a downstream dependency fails.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
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
func (f *sftpFile) Stat() (os.FileInfo, error) {
	return f.f.Stat()
}

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
func (f *sftpFile) Sync() error {
	return nil
}

// Truncate changes the size of the file.
//
// Summary: Truncate changes the size of the file.
//
// Parameters:
//   - size (int64): The numeric value for size.
//
// Returns:
//   - error: An error if the execution fails, otherwise nil.
//
// Errors:
//   - Returns an error if the operation fails, invalid input is provided, or a downstream dependency fails.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
func (f *sftpFile) Truncate(size int64) error {
	return f.f.Truncate(size)
}

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
func (f *sftpFile) WriteString(s string) (ret int, err error) {
	return f.f.Write([]byte(s))
}
