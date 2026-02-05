// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/pkg/sftp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
	"google.golang.org/protobuf/proto"
)

// Helper to start a local SFTP server
func startSFTPServer(t *testing.T, authorizedKey ssh.PublicKey) (string, *ssh.ServerConfig, func()) {
	// 1. Generate Host Key
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	keyBytes := x509.MarshalPKCS1PrivateKey(key)
	keyPem := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: keyBytes,
	})
	signer, err := ssh.ParsePrivateKey(keyPem)
	require.NoError(t, err)

	config := &ssh.ServerConfig{
		PasswordCallback: func(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
			if c.User() == "testuser" && string(pass) == "testpass" {
				return nil, nil
			}
			return nil, fmt.Errorf("password rejected for %q", c.User())
		},
		PublicKeyCallback: func(c ssh.ConnMetadata, pubKey ssh.PublicKey) (*ssh.Permissions, error) {
			if authorizedKey != nil {
				// We need to compare marshaled bytes to be safe
				if ssh.FingerprintSHA256(pubKey) == ssh.FingerprintSHA256(authorizedKey) {
					return nil, nil
				}
			}
			return nil, fmt.Errorf("unknown public key for %q", c.User())
		},
	}
	config.AddHostKey(signer)

	// 2. Listen on random port
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	addr := listener.Addr().String()

	// 3. Accept connections
	go func() {
		for {
			nConn, err := listener.Accept()
			if err != nil {
				return
			}

			go func(conn net.Conn) {
				_, chans, reqs, err := ssh.NewServerConn(conn, config)
				if err != nil {
					return
				}
				go ssh.DiscardRequests(reqs)

				for newChannel := range chans {
					if newChannel.ChannelType() != "session" {
						newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
						continue
					}
					channel, requests, err := newChannel.Accept()
					if err != nil {
						continue
					}

					go func(in <-chan *ssh.Request) {
						for req := range in {
							ok := false
							switch req.Type {
							case "subsystem":
								if string(req.Payload[4:]) == "sftp" {
									ok = true
									go func() {
										defer channel.Close()
										server, err := sftp.NewServer(
											channel,
										)
										if err != nil {
											return
										}
										if err := server.Serve(); err == io.EOF {
											server.Close()
										}
									}()
								}
							}
							req.Reply(ok, nil)
						}
					}(requests)
				}
			}(nConn)
		}
	}()

	return addr, config, func() { listener.Close() }
}

func TestSftpProvider(t *testing.T) {
	// Create a temporary directory for the SFTP server to serve
	addr, _, cleanup := startSFTPServer(t, nil)
	defer cleanup()

	// Create temp dir for testing
	tmpDir, err := os.MkdirTemp("", "sftp-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	config := configv1.SftpFs_builder{
		Address:  proto.String(addr),
		Username: proto.String("testuser"),
		Password: proto.String("testpass"),
	}.Build()

	provider, err := NewSftpProvider(config)
	require.NoError(t, err)
	defer provider.Close()

	fs := provider.GetFs()

	t.Run("Create and WriteString", func(t *testing.T) {
		filename := filepath.Join(tmpDir, "test.txt")
		f, err := fs.Create(filename)
		require.NoError(t, err)

		_, err = f.WriteString("Hello SFTP")
		require.NoError(t, err)

		err = f.Close()
		require.NoError(t, err)

		// Verify content
		content, err := os.ReadFile(filename)
		require.NoError(t, err)
		assert.Equal(t, "Hello SFTP", string(content))
	})

	t.Run("Open and Read", func(t *testing.T) {
		filename := filepath.Join(tmpDir, "test.txt")
		f, err := fs.Open(filename)
		require.NoError(t, err)
		defer f.Close()

		content, err := io.ReadAll(f)
		require.NoError(t, err)
		assert.Equal(t, "Hello SFTP", string(content))
	})

	t.Run("Stat", func(t *testing.T) {
		filename := filepath.Join(tmpDir, "test.txt")
		info, err := fs.Stat(filename)
		require.NoError(t, err)
		assert.Equal(t, "test.txt", info.Name())
		assert.Equal(t, int64(10), info.Size())
	})

	t.Run("ResolvePath", func(t *testing.T) {
		path, err := provider.ResolvePath("/foo/bar")
		require.NoError(t, err)
		assert.Equal(t, "/foo/bar", path)
	})

	t.Run("Mkdir and Remove", func(t *testing.T) {
		dir := filepath.Join(tmpDir, "subdir")
		err := fs.Mkdir(dir, 0755)
		require.NoError(t, err)

		info, err := fs.Stat(dir)
		require.NoError(t, err)
		assert.True(t, info.IsDir())

		err = fs.Remove(dir)
		require.NoError(t, err)

		_, err = fs.Stat(dir)
		assert.Error(t, err)
		assert.True(t, os.IsNotExist(err))
	})

	t.Run("MkdirAll and RemoveAll", func(t *testing.T) {
		dir := filepath.Join(tmpDir, "nested/sub/dir")
		err := fs.MkdirAll(dir, 0755)
		require.NoError(t, err)

		err = fs.RemoveAll(filepath.Join(tmpDir, "nested"))
		require.NoError(t, err)

		_, err = fs.Stat(filepath.Join(tmpDir, "nested"))
		assert.Error(t, err)
	})

	t.Run("Rename", func(t *testing.T) {
		src := filepath.Join(tmpDir, "src.txt")
		dst := filepath.Join(tmpDir, "dst.txt")
		f, err := fs.Create(src)
		require.NoError(t, err)
		f.Close()

		err = fs.Rename(src, dst)
		require.NoError(t, err)

		_, err = fs.Stat(src)
		assert.Error(t, err)
		_, err = fs.Stat(dst)
		require.NoError(t, err)
	})

	t.Run("Readdir", func(t *testing.T) {
		dir := filepath.Join(tmpDir, "readdir_test")
		err := fs.Mkdir(dir, 0755)
		require.NoError(t, err)

		f1, _ := fs.Create(filepath.Join(dir, "f1.txt"))
		f1.Close()
		f2, _ := fs.Create(filepath.Join(dir, "f2.txt"))
		f2.Close()

		d, err := fs.Open(dir)
		require.NoError(t, err)
		defer d.Close()

		infos, err := d.Readdir(0)
		require.NoError(t, err)
		assert.Len(t, infos, 2)
	})

	t.Run("ReadAt and Seek", func(t *testing.T) {
		path := filepath.Join(tmpDir, "seek.txt")
		f, err := fs.Create(path)
		require.NoError(t, err)
		_, _ = f.WriteString("0123456789")
		f.Close()

		f, err = fs.Open(path)
		require.NoError(t, err)
		defer f.Close()

		// ReadAt
		buf := make([]byte, 5)
		n, err := f.ReadAt(buf, 2)
		require.NoError(t, err)
		assert.Equal(t, 5, n)
		assert.Equal(t, "23456", string(buf))

		// Seek
		off, err := f.Seek(5, io.SeekStart)
		require.NoError(t, err)
		assert.Equal(t, int64(5), off)

		buf2 := make([]byte, 5)
		n, err = f.Read(buf2)
		require.NoError(t, err)
		assert.Equal(t, 5, n)
		assert.Equal(t, "56789", string(buf2))
	})

	t.Run("WriteAt", func(t *testing.T) {
		path := filepath.Join(tmpDir, "writeat.txt")
		f, err := fs.Create(path)
		require.NoError(t, err)

		_, err = f.WriteString("00000")
		require.NoError(t, err)

		n, err := f.WriteAt([]byte("11"), 1)
		require.NoError(t, err)
		assert.Equal(t, 2, n)

		f.Close()

		content, _ := os.ReadFile(path)
		assert.Equal(t, "01100", string(content))
	})

	t.Run("Write and Name", func(t *testing.T) {
		path := filepath.Join(tmpDir, "write.txt")
		f, err := fs.Create(path)
		require.NoError(t, err)

		n, err := f.Write([]byte("foo"))
		require.NoError(t, err)
		assert.Equal(t, 3, n)

		assert.Equal(t, "write.txt", filepath.Base(f.Name()))

		f.Close()

		content, _ := os.ReadFile(path)
		assert.Equal(t, "foo", string(content))
	})

	t.Run("Chown", func(t *testing.T) {
		path := filepath.Join(tmpDir, "chown.txt")
		f, _ := fs.Create(path)
		f.Close()

		// Just call it, expecting either nil or error, but covering the line.
		_ = fs.Chown(path, 1000, 1000)
	})
}

func TestNewSftpProvider_KeyAuth(t *testing.T) {
	// Generate a key pair for the client
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	// Get public key for server
	publicKey, err := ssh.NewPublicKey(&privateKey.PublicKey)
	require.NoError(t, err)

	addr, _, cleanup := startSFTPServer(t, publicKey)
	defer cleanup()

	// Write private key to temp file
	keyDir, err := os.MkdirTemp("", "keys")
	require.NoError(t, err)
	defer os.RemoveAll(keyDir)

	keyFile := filepath.Join(keyDir, "id_rsa")
	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})
	err = os.WriteFile(keyFile, keyPEM, 0600)
	require.NoError(t, err)

	config := configv1.SftpFs_builder{
		Address:  proto.String(addr),
		Username: proto.String("testuser"),
		KeyPath:  proto.String(keyFile),
	}.Build()

	provider, err := NewSftpProvider(config)
	require.NoError(t, err)
	defer provider.Close()

	// Verify connection works
	_, err = provider.GetFs().Stat(".")
	require.NoError(t, err)
}

func TestNewSftpProvider_Errors(t *testing.T) {
	t.Run("Nil Config", func(t *testing.T) {
		_, err := NewSftpProvider(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "sftp config is nil")
	})

	t.Run("Missing Key File", func(t *testing.T) {
		config := configv1.SftpFs_builder{
			Username: proto.String("user"),
			KeyPath:  proto.String("/non/existent/key"),
		}.Build()
		_, err := NewSftpProvider(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read private key")
	})

	t.Run("Invalid Key File", func(t *testing.T) {
		keyDir, err := os.MkdirTemp("", "badkeys")
		require.NoError(t, err)
		defer os.RemoveAll(keyDir)

		keyFile := filepath.Join(keyDir, "bad_key")
		err = os.WriteFile(keyFile, []byte("not a key"), 0600)
		require.NoError(t, err)

		config := configv1.SftpFs_builder{
			Username: proto.String("user"),
			KeyPath:  proto.String(keyFile),
		}.Build()
		_, err = NewSftpProvider(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse private key")
	})

	t.Run("Dial Error", func(t *testing.T) {
		// Try to find a free port that is NOT listening
		l, err := net.Listen("tcp", "127.0.0.1:0")
		require.NoError(t, err)
		freeAddr := l.Addr().String()
		l.Close()

		config := configv1.SftpFs_builder{
			Address:  proto.String(freeAddr),
			Username: proto.String("user"),
			Password: proto.String("pass"),
		}.Build()

		_, err = NewSftpProvider(config)
		assert.Error(t, err)
	})
}
