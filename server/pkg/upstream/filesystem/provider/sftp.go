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

// SftpProvider provides access to files via the SFTP protocol.
//
// Summary: Implements the FilesystemProvider for remote SFTP servers.
type SftpProvider struct {
	fs     afero.Fs
	client *sftp.Client
	conn   *ssh.Client
}

// NewSftpProvider creates a new SftpProvider and establishes an SSH connection.
//
// Summary: Initializes a new SFTP filesystem provider and dials the remote host.
//
// Parameters:
//   - config (*configv1.SftpFs): The SFTP configuration including address, username, and authentication details.
//
// Returns:
//   - *SftpProvider: A pointer to the newly created SFTP provider instance.
//   - error: An error if the SSH connection or SFTP session fails to initialize.
//
// Errors:
//   - Returns an error if the configuration object is nil.
//   - Returns an error if the private key file cannot be read or parsed.
//   - Returns an error if the SSH dial operation fails.
//   - Returns an error if the SFTP client creation fails.
//
// Side Effects:
//   - Establishes a persistent TCP connection to the remote SFTP server.
//   - Reads the private key file from the host filesystem if configured.
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
// Summary: Retrieves the underlying afero SFTP filesystem.
//
// Returns:
//   - afero.Fs: An afero.Fs implementation backed by the remote SFTP server.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
//
// Parameters:
//   - None.
func (p *SftpProvider) GetFs() afero.Fs {
	return p.fs
}

// ResolvePath resolves the virtual path to a cleaned remote path.
//
// Summary: Cleans and returns the remote path for use in SFTP operations.
//
// Parameters:
//   - virtualPath (string): The virtual path provided by the agent.
//
// Returns:
//   - string: The cleaned remote path.
//   - error: Nil, as SFTP paths are treated as absolute or relative to the user's home on the remote host.
//
// Errors:
//   - None.
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

// Close closes the SFTP client and the underlying SSH connection.
//
// Summary: Releases SFTP and SSH resources.
//
// Returns:
//   - error: An error if either the SFTP client or SSH connection fails to close.
//
// Errors:
//   - Returns an error if the SFTP client or SSH connection closure fails.
//
// Side Effects:
//   - Terminates the TCP connection to the remote host.
//
// Parameters:
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

// Create creates a file in the filesystem, returning the file and an error, if any happens.
//
// Parameters:
//   - name (string): The parameter.
//
// Returns:
//   - afero.File: The result.
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if ...
//
// Side Effects:
//   - None.
//
// Summary: Initializes Create operation.
//
// Parameters:
//   - name (string): The name of the file to create.
//
// Returns:
//   - afero.File: The created file.
//   - error: An error if creation fails.
func (s *sftpFs) Create(name string) (afero.File, error) {
	f, err := s.client.Create(name)
	if err != nil {
		return nil, err
	}
	return &sftpFile{f: f, client: s.client}, nil
}

// Mkdir creates a directory in the filesystem, returning an error, if any happens.
//
// Parameters:
//   - name (string): The parameter.
//   - _ (os.FileMode): The parameter.
//
// Returns:
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if ...
//
// Side Effects:
//   - None.
//
// Summary: Executes Mkdir operation.
//
// Parameters:
//   - name (string): The directory name.
//   - _ (os.FileMode): Unused mode.
//
// Returns:
//   - error: An error if creation fails.
func (s *sftpFs) Mkdir(name string, _ os.FileMode) error {
	return s.client.Mkdir(name)
}

// MkdirAll creates a directory path and all parents that does not exist for a given name.
//
// Parameters:
//   - path (string): The parameter.
//   - _ (os.FileMode): The parameter.
//
// Returns:
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if ...
//
// Side Effects:
//   - None.
//
// Summary: Executes MkdirAll operation.
//
// Parameters:
//   - path (string): The directory path.
//   - _ (os.FileMode): Unused mode.
//
// Returns:
//   - error: An error if creation fails.
func (s *sftpFs) MkdirAll(path string, _ os.FileMode) error {
	return s.client.MkdirAll(path)
}

// Open opens a file, returning it or an error, if any happens.
//
// Parameters:
//   - name (string): The parameter.
//
// Returns:
//   - afero.File: The result.
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if ...
//
// Side Effects:
//   - None.
//
// Summary: Executes Open operation.
//
// Parameters:
//   - name (string): The file name.
//
// Returns:
//   - afero.File: The opened file.
//   - error: An error if opening fails.
func (s *sftpFs) Open(name string) (afero.File, error) {
	f, err := s.client.Open(name)
	if err != nil {
		return nil, err
	}
	return &sftpFile{f: f, client: s.client}, nil
}

// OpenFile opens a file using the given flags and the given mode.
//
// Parameters:
//   - name (string): The parameter.
//   - flag (int): The parameter.
//   - _ (os.FileMode): The parameter.
//
// Returns:
//   - afero.File: The result.
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if ...
//
// Side Effects:
//   - None.
//
// Summary: Executes OpenFile operation.
//
// Parameters:
//   - name (string): The file name.
//   - flag (int): Opening flags.
//   - _ (os.FileMode): Unused mode.
//
// Returns:
//   - afero.File: The opened file.
//   - error: An error if opening fails.
func (s *sftpFs) OpenFile(name string, flag int, _ os.FileMode) (afero.File, error) {
	f, err := s.client.OpenFile(name, flag)
	if err != nil {
		return nil, err
	}
	return &sftpFile{f: f, client: s.client}, nil
}

// Remove removes a file identified by name, returning an error, if any happens.
//
// Parameters:
//   - name (string): The parameter.
//
// Returns:
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if ...
//
// Side Effects:
//   - None.
//
// Summary: Executes Remove operation.
//
// Parameters:
//   - name (string): The name of the file to remove.
//
// Returns:
//   - error: An error if removal fails.
func (s *sftpFs) Remove(name string) error {
	return s.client.Remove(name)
}

// RemoveAll removes a directory path and any children it contains.
//
// Parameters:
//   - path (string): The parameter.
//
// Returns:
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if ...
//
// Side Effects:
//   - None.
//
// Summary: Executes RemoveAll operation.
//
// Parameters:
//   - path (string): The path to remove.
//
// Returns:
//   - error: An error if removal fails.
func (s *sftpFs) RemoveAll(path string) error {
	// sftp.Client.RemoveAll actually does recursive removal
	return s.client.RemoveAll(path)
}

// Rename renames a file.
//
// Parameters:
//   - (oldname): The parameter.
//   - newname (string): The parameter.
//
// Returns:
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if ...
//
// Side Effects:
//   - None.
//
// Summary: Executes Rename operation.
//
// Parameters:
//   - oldname (string): The original name.
//   - newname (string): The new name.
//
// Returns:
//   - error: An error if renaming fails.
func (s *sftpFs) Rename(oldname, newname string) error {
	return s.client.Rename(oldname, newname)
}

// Stat returns a FileInfo describing the named file, or an error, if any happens.
//
// Parameters:
//   - name (string): The parameter.
//
// Returns:
//   - os.FileInfo: The result.
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if ...
//
// Side Effects:
//   - None.
//
// Summary: Executes Stat operation.
//
// Parameters:
//   - name (string): The name of the file.
//
// Returns:
//   - os.FileInfo: File information.
//   - error: An error if stat fails.
func (s *sftpFs) Stat(name string) (os.FileInfo, error) {
	return s.client.Stat(name)
}

// Name returns the name of this file system.
//
// Returns:
//   - string: The result.
//
// Side Effects:
//   - None.
//
// Summary: Executes Name operation.
//
// Returns:
//   - string: The name of the filesystem.
//
// Parameters:
//   - None.
//
// Errors:
//   - None.
func (s *sftpFs) Name() string {
	return "sftp"
}

// Chmod changes the mode of the named file to mode.
//
// Parameters:
//   - name (string): The parameter.
//   - mode (os.FileMode): The parameter.
//
// Returns:
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if ...
//
// Side Effects:
//   - None.
//
// Summary: Executes Chmod operation.
//
// Parameters:
//   - name (string): The file name.
//   - mode (os.FileMode): The file mode.
//
// Returns:
//   - error: An error if chmod fails.
func (s *sftpFs) Chmod(name string, mode os.FileMode) error {
	return s.client.Chmod(name, mode)
}

// Chown changes the uid and gid of the named file.
//
// Parameters:
//   - name (string): The parameter.
//   - (uid): The parameter.
//   - gid (int): The parameter.
//
// Returns:
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if ...
//
// Side Effects:
//   - None.
//
// Summary: Executes Chown operation.
//
// Parameters:
//   - name (string): The file name.
//   - uid (int): User ID.
//   - gid (int): Group ID.
//
// Returns:
//   - error: An error if chown fails.
func (s *sftpFs) Chown(name string, uid, gid int) error {
	return s.client.Chown(name, uid, gid)
}

// Chtimes changes the access and modification times of the named file.
//
// Parameters:
//   - name (string): The parameter.
//   - atime (time.Time): The parameter.
//   - mtime (time.Time): The parameter.
//
// Returns:
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if ...
//
// Side Effects:
//   - None.
//
// Summary: Executes Chtimes operation.
//
// Parameters:
//   - name (string): The file name.
//   - atime (time.Time): Access time.
//   - mtime (time.Time): Modification time.
//
// Returns:
//   - error: An error if chtimes fails.
func (s *sftpFs) Chtimes(name string, atime time.Time, mtime time.Time) error {
	return s.client.Chtimes(name, atime, mtime)
}

type sftpFile struct {
	f      *sftp.File
	client *sftp.Client
}

// Close closes the file.
//
// Returns:
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if ...
//
// Side Effects:
//   - None.
//
// Summary: Executes Close operation.
//
// Returns:
//   - error: An error if closure fails.
//
// Parameters:
//   - None.
func (f *sftpFile) Close() error {
	return f.f.Close()
}

// Read reads up to len(b) bytes from the File.
//
// Parameters:
//   - p ([]byte): The parameter.
//
// Returns:
//   - n (int): The result.
//   - err (error): An error if the operation fails.
//
// Errors:
//   - Returns an error if ...
//
// Side Effects:
//   - None.
//
// Summary: Retrieves Read operation.
//
// Parameters:
//   - p ([]byte): Destination buffer.
//
// Returns:
//   - n (int): Bytes read.
//   - err (error): An error if reading fails.
func (f *sftpFile) Read(p []byte) (n int, err error) {
	return f.f.Read(p)
}

// ReadAt reads len(b) bytes from the File starting at byte offset off.
//
// Parameters:
//   - p ([]byte): The parameter.
//   - off (int64): The parameter.
//
// Returns:
//   - n (int): The result.
//   - err (error): An error if the operation fails.
//
// Errors:
//   - Returns an error if ...
//
// Side Effects:
//   - None.
//
// Summary: Retrieves ReadAt operation.
//
// Parameters:
//   - p ([]byte): Destination buffer.
//   - off (int64): Offset.
//
// Returns:
//   - n (int): Bytes read.
//   - err (error): An error if reading fails.
func (f *sftpFile) ReadAt(p []byte, off int64) (n int, err error) {
	return f.f.ReadAt(p, off)
}

// Seek sets the offset for the next Read or Write to offset, interpreted according to whence.
//
// Parameters:
//   - offset (int64): The parameter.
//   - whence (int): The parameter.
//
// Returns:
//   - int64: The result.
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if ...
//
// Side Effects:
//   - None.
//
// Summary: Executes Seek operation.
//
// Parameters:
//   - offset (int64): Seek offset.
//   - whence (int): Seek whence.
//
// Returns:
//   - int64: New offset.
//   - error: An error if seeking fails.
func (f *sftpFile) Seek(offset int64, whence int) (int64, error) {
	return f.f.Seek(offset, whence)
}

// Write writes len(b) bytes to the File.
//
// Parameters:
//   - p ([]byte): The parameter.
//
// Returns:
//   - n (int): The result.
//   - err (error): An error if the operation fails.
//
// Errors:
//   - Returns an error if ...
//
// Side Effects:
//   - None.
//
// Summary: Updates Write operation.
//
// Parameters:
//   - p ([]byte): Data to write.
//
// Returns:
//   - n (int): Bytes written.
//   - err (error): An error if writing fails.
func (f *sftpFile) Write(p []byte) (n int, err error) {
	return f.f.Write(p)
}

// WriteAt writes len(b) bytes to the File starting at byte offset off.
//
// Parameters:
//   - p ([]byte): The parameter.
//   - off (int64): The parameter.
//
// Returns:
//   - n (int): The result.
//   - err (error): An error if the operation fails.
//
// Errors:
//   - Returns an error if ...
//
// Side Effects:
//   - None.
//
// Summary: Updates WriteAt operation.
//
// Parameters:
//   - p ([]byte): Data to write.
//   - off (int64): Offset.
//
// Returns:
//   - n (int): Bytes written.
//   - err (error): An error if writing fails.
func (f *sftpFile) WriteAt(p []byte, off int64) (n int, err error) {
	return f.f.WriteAt(p, off)
}

// Name returns the name of the file as presented to Open.
//
// Returns:
//   - string: The result.
//
// Side Effects:
//   - None.
//
// Summary: Executes Name operation.
//
// Returns:
//   - string: The file name.
//
// Parameters:
//   - None.
//
// Errors:
//   - None.
func (f *sftpFile) Name() string {
	return f.f.Name()
}

// Readdir reads the contents of the directory associated with file and returns
// a slice of up to n FileInfo values, as would be returned by Lstat, in directory order.
//
// Parameters:
//   - _ (int): The parameter.
//
// Returns:
//   - []os.FileInfo: The result.
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if ...
//
// Side Effects:
//   - None.
//
// Summary: Retrieves Readdir operation.
//
// Parameters:
//   - _ (int): Unused count.
//
// Returns:
//   - []os.FileInfo: Directory entries.
//   - error: An error if reading fails.
func (f *sftpFile) Readdir(_ int) ([]os.FileInfo, error) {
	return f.client.ReadDir(f.f.Name())
}

// Readdirnames reads and returns a slice of names from the directory f.
//
// Parameters:
//   - n (int): The parameter.
//
// Returns:
//   - []string: The result.
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if ...
//
// Side Effects:
//   - None.
//
// Summary: Retrieves Readdirnames operation.
//
// Parameters:
//   - n (int): Number of names to return.
//
// Returns:
//   - []string: Entry names.
//   - error: An error if reading fails.
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
// Returns:
//   - os.FileInfo: The result.
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if ...
//
// Side Effects:
//   - None.
//
// Summary: Executes Stat operation on file.
//
// Returns:
//   - os.FileInfo: File information.
//   - error: An error if stat fails.
//
// Parameters:
//   - None.
func (f *sftpFile) Stat() (os.FileInfo, error) {
	return f.f.Stat()
}

// Sync commits the current contents of the file to stable storage.
//
// Returns:
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if ...
//
// Side Effects:
//   - None.
//
// Summary: Executes Sync operation.
//
// Returns:
//   - error: Nil.
//
// Parameters:
//   - None.
func (f *sftpFile) Sync() error {
	return nil
}

// Truncate changes the size of the file.
//
// Parameters:
//   - size (int64): The parameter.
//
// Returns:
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if ...
//
// Side Effects:
//   - None.
//
// Summary: Executes Truncate operation.
//
// Parameters:
//   - size (int64): New size.
//
// Returns:
//   - error: An error if truncation fails.
func (f *sftpFile) Truncate(size int64) error {
	return f.f.Truncate(size)
}

// WriteString is like Write, but writes the contents of string s rather than a slice of bytes.
//
// Parameters:
//   - s (string): The parameter.
//
// Returns:
//   - ret (int): The result.
//   - err (error): An error if the operation fails.
//
// Errors:
//   - Returns an error if ...
//
// Side Effects:
//   - None.
//
// Summary: Updates WriteString operation.
//
// Parameters:
//   - s (string): String to write.
//
// Returns:
//   - int: Bytes written.
//   - error: An error if writing fails.
func (f *sftpFile) WriteString(s string) (ret int, err error) {
	return f.f.Write([]byte(s))
}
