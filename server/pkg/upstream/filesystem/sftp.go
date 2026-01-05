// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package filesystem

import (
	"fmt"
	"os"
	"strings"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/pkg/sftp"
	"github.com/spf13/afero"
	"golang.org/x/crypto/ssh"
)

func (u *Upstream) createSftpFilesystem(config *configv1.SftpFs) (afero.Fs, error) {
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

	u.mu.Lock()
	u.closers = append(u.closers, client, conn)
	u.mu.Unlock()

	return &sftpFs{client: client}, nil
}

type sftpFs struct {
	client *sftp.Client
}

// Create creates a file in the filesystem, returning the file and an error, if any happens.
func (s *sftpFs) Create(name string) (afero.File, error) {
	f, err := s.client.Create(name)
	if err != nil {
		return nil, err
	}
	return &sftpFile{f: f, client: s.client}, nil
}

// Mkdir creates a directory in the filesystem, returning an error, if any happens.
func (s *sftpFs) Mkdir(name string, _ os.FileMode) error {
	return s.client.Mkdir(name)
}

// MkdirAll creates a directory path and all parents that does not exist for a given name.
func (s *sftpFs) MkdirAll(path string, _ os.FileMode) error {
	return s.client.MkdirAll(path)
}

// Open opens a file, returning it or an error, if any happens.
func (s *sftpFs) Open(name string) (afero.File, error) {
	f, err := s.client.Open(name)
	if err != nil {
		return nil, err
	}
	return &sftpFile{f: f, client: s.client}, nil
}

// OpenFile opens a file using the given flags and the given mode.
func (s *sftpFs) OpenFile(name string, flag int, _ os.FileMode) (afero.File, error) {
	f, err := s.client.OpenFile(name, flag)
	if err != nil {
		return nil, err
	}
	return &sftpFile{f: f, client: s.client}, nil
}

// Remove removes a file identified by name, returning an error, if any happens.
func (s *sftpFs) Remove(name string) error {
	return s.client.Remove(name)
}

// RemoveAll removes a directory path and any children it contains.
func (s *sftpFs) RemoveAll(path string) error {
	return s.client.Remove(path)
}

// Rename renames a file.
func (s *sftpFs) Rename(oldname, newname string) error {
	return s.client.Rename(oldname, newname)
}

// Stat returns a FileInfo describing the named file, or an error, if any happens.
func (s *sftpFs) Stat(name string) (os.FileInfo, error) {
	return s.client.Stat(name)
}

// Name returns the name of this file system.
func (s *sftpFs) Name() string {
	return "sftp"
}

// Chmod changes the mode of the named file to mode.
func (s *sftpFs) Chmod(name string, mode os.FileMode) error {
	return s.client.Chmod(name, mode)
}

// Chown changes the uid and gid of the named file.
func (s *sftpFs) Chown(name string, uid, gid int) error {
	return s.client.Chown(name, uid, gid)
}

// Chtimes changes the access and modification times of the named file.
func (s *sftpFs) Chtimes(name string, atime time.Time, mtime time.Time) error {
	return s.client.Chtimes(name, atime, mtime)
}

type sftpFile struct {
	f      *sftp.File
	client *sftp.Client
}

// Close closes the file.
func (f *sftpFile) Close() error {
	return f.f.Close()
}

// Read reads up to len(b) bytes from the File.
func (f *sftpFile) Read(p []byte) (n int, err error) {
	return f.f.Read(p)
}

// ReadAt reads len(b) bytes from the File starting at byte offset off.
func (f *sftpFile) ReadAt(p []byte, off int64) (n int, err error) {
	return f.f.ReadAt(p, off)
}

// Seek sets the offset for the next Read or Write to offset, interpreted according to whence.
func (f *sftpFile) Seek(offset int64, whence int) (int64, error) {
	return f.f.Seek(offset, whence)
}

// Write writes len(b) bytes to the File.
func (f *sftpFile) Write(p []byte) (n int, err error) {
	return f.f.Write(p)
}

// WriteAt writes len(b) bytes to the File starting at byte offset off.
func (f *sftpFile) WriteAt(p []byte, off int64) (n int, err error) {
	return f.f.WriteAt(p, off)
}

// Name returns the name of the file as presented to Open.
func (f *sftpFile) Name() string {
	return f.f.Name()
}

// Readdir reads the contents of the directory associated with file and returns
// a slice of up to n FileInfo values, as would be returned by Lstat, in directory order.
func (f *sftpFile) Readdir(_ int) ([]os.FileInfo, error) {
	return f.client.ReadDir(f.f.Name())
}

// Readdirnames reads and returns a slice of names from the directory f.
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
func (f *sftpFile) Stat() (os.FileInfo, error) {
	return f.f.Stat()
}

// Sync commits the current contents of the file to stable storage.
func (f *sftpFile) Sync() error {
	return nil
}

// Truncate changes the size of the file.
func (f *sftpFile) Truncate(size int64) error {
	return f.f.Truncate(size)
}

// WriteString is like Write, but writes the contents of string s rather than a slice of bytes.
func (f *sftpFile) WriteString(s string) (ret int, err error) {
	return f.f.Write([]byte(s))
}
