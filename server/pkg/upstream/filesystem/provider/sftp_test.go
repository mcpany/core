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
)

func ptr(s string) *string {
	return &s
}

// Helper to start a local SFTP server
func startSFTPServer(t *testing.T) (string, *ssh.ServerConfig, func()) {
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
	// Create a temporary directory for the SFTP server to serve (if we could restrict it easily)
	// Note: pkg/sftp server serves the local filesystem of the process.
	// So we should operate in a temp dir.

	addr, _, cleanup := startSFTPServer(t)
	defer cleanup()

	// Create temp dir for testing
	tmpDir, err := os.MkdirTemp("", "sftp-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	config := &configv1.SftpFs{
		Address:  &addr,
		Username: ptr("testuser"),
		Password: ptr("testpass"),
	}

	provider, err := NewSftpProvider(config)
	require.NoError(t, err)
	defer provider.Close()

	fs := provider.GetFs()

	t.Run("Create and Write", func(t *testing.T) {
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
}
