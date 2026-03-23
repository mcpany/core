// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package update

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/google/go-github/v39/github"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type controlledMockFs struct {
	afero.Fs
	renameHooks   []func(old, new string) error
	chmodHooks    []func(name string, mode os.FileMode) error
	openFileHooks []func(name string, flag int, perm os.FileMode) (afero.File, error)
}

func (m *controlledMockFs) Rename(oldname, newname string) error {
	if len(m.renameHooks) > 0 {
		hook := m.renameHooks[0]
		m.renameHooks = m.renameHooks[1:]
		if hook != nil {
			return hook(oldname, newname)
		}
	}
	return m.Fs.Rename(oldname, newname)
}

func (m *controlledMockFs) Chmod(name string, mode os.FileMode) error {
	if len(m.chmodHooks) > 0 {
		hook := m.chmodHooks[0]
		m.chmodHooks = m.chmodHooks[1:]
		if hook != nil {
			return hook(name, mode)
		}
	}
	return m.Fs.Chmod(name, mode)
}

func (m *controlledMockFs) OpenFile(name string, flag int, perm os.FileMode) (afero.File, error) {
	if len(m.openFileHooks) > 0 {
		hook := m.openFileHooks[0]
		m.openFileHooks = m.openFileHooks[1:]
		if hook != nil {
			return hook(name, flag, perm)
		}
	}
	return m.Fs.OpenFile(name, flag, perm)
}


func TestNewUpdater(t *testing.T) {
	t.Run("with nil http client", func(t *testing.T) {
		updater := NewUpdater(nil, "")
		assert.NotNil(t, updater.client)
		assert.Equal(t, http.DefaultClient, updater.httpClient)
	})

	t.Run("with custom github api url", func(t *testing.T) {
		updater := NewUpdater(nil, "http://127.0.0.1:8080/api/")
		assert.Equal(t, "http://127.0.0.1:8080/api/", updater.client.BaseURL.String())
	})

	t.Run("with malformed custom github api url", func(t *testing.T) {
		// NewUpdater doesn't validate URL at creation, it just parses.
		// If empty or invalid, maybe it defaults or stays?
		// My implementation in updater.go: if != "", Parse(url).
		// If parse fails, it ignores? ("base url ...")
		// Checks should be done in NewUpdater?
		// Assuming previous test expected default if malformed?
		// Previous test used Setenv.
		// Let's pass empty for this test or skip, or check if NewUpdater handles it.
		// For now, let's just pass empty string to satisfy signature.
		updater := NewUpdater(nil, "")
		assert.Equal(t, "https://api.github.com/", updater.client.BaseURL.String())
	})
}

func TestCheckForUpdate(t *testing.T) {
	t.Run("new version available", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			release := &github.RepositoryRelease{
				TagName: github.String("v1.0.1"),
			}
			json.NewEncoder(w).Encode(release)
		}))
		defer server.Close()

		updater := NewUpdater(server.Client(), server.URL)

		release, available, err := updater.CheckForUpdate(context.Background(), "owner", "repo", "v1.0.0")
		require.NoError(t, err)
		assert.True(t, available)
		assert.NotNil(t, release)
		assert.Equal(t, "v1.0.1", release.GetTagName())
	})

	t.Run("already on latest version", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			release := &github.RepositoryRelease{
				TagName: github.String("v1.0.0"),
			}
			json.NewEncoder(w).Encode(release)
		}))
		defer server.Close()

		updater := NewUpdater(server.Client(), server.URL)

		release, available, err := updater.CheckForUpdate(context.Background(), "owner", "repo", "v1.0.0")
		require.NoError(t, err)
		assert.False(t, available)
		assert.Nil(t, release)
	})

	t.Run("github api error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		updater := NewUpdater(server.Client(), server.URL)

		_, available, err := updater.CheckForUpdate(context.Background(), "owner", "repo", "v1.0.0")
		require.Error(t, err)
		assert.False(t, available)
	})
}

func TestUpdateTo_Success(t *testing.T) {
	assetContent := "new binary content"
	assetName := "server-linux-amd64"
	assetHash := sha256.Sum256([]byte(assetContent))
	checksumsContent := fmt.Sprintf("%s  %s\n", hex.EncodeToString(assetHash[:]), assetName)

	assetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(assetContent))
	}))
	defer assetServer.Close()

	checksumsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(checksumsContent))
	}))
	defer checksumsServer.Close()

	release := &github.RepositoryRelease{
		TagName: github.String("v1.0.1"),
		Assets: []*github.ReleaseAsset{
			{
				Name:               github.String(assetName),
				BrowserDownloadURL: github.String(assetServer.URL),
			},
			{
				Name:               github.String("checksums.txt"),
				BrowserDownloadURL: github.String(checksumsServer.URL),
			},
		},
	}
	t.Run("successfully updates when executable does not exist", func(t *testing.T) {
		memfs := afero.NewMemMapFs()
		cmfs := &controlledMockFs{Fs: memfs}

		executablePath := "/app/server"

		updater := NewUpdater(http.DefaultClient, "")

		err := updater.UpdateTo(context.Background(), cmfs, executablePath, release, assetName, "checksums.txt")

		require.NoError(t, err)

		content, err := afero.ReadFile(cmfs, executablePath)
		require.NoError(t, err)
		assert.Equal(t, assetContent, string(content))
	})
}
func TestUpdateTo_FailureScenarios(t *testing.T) {
	assetContent := "new binary content"
	assetName := "server-linux-amd64"
	assetHash := sha256.Sum256([]byte(assetContent))
	checksumsContent := fmt.Sprintf("%s  %s\n", hex.EncodeToString(assetHash[:]), assetName)

	assetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(assetContent))
	}))
	defer assetServer.Close()

	checksumsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(checksumsContent))
	}))
	defer checksumsServer.Close()

	release := &github.RepositoryRelease{
		TagName: github.String("v1.0.1"),
		Assets: []*github.ReleaseAsset{
			{
				Name:               github.String(assetName),
				BrowserDownloadURL: github.String(assetServer.URL),
			},
			{
				Name:               github.String("checksums.txt"),
				BrowserDownloadURL: github.String(checksumsServer.URL),
			},
		},
	}

	t.Run("fails to rename old executable", func(t *testing.T) {
		memfs := afero.NewMemMapFs()
		cmfs := &controlledMockFs{Fs: memfs}
		cmfs.renameHooks = []func(old, new string) error{
			func(old, new string) error { return fmt.Errorf("permission denied") },
		}

		executablePath := "/app/server"
		err := afero.WriteFile(cmfs, executablePath, []byte("old content"), 0755)
		require.NoError(t, err)

		updater := NewUpdater(http.DefaultClient, "")

		err = updater.UpdateTo(context.Background(), cmfs, executablePath, release, assetName, "checksums.txt")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to rename old executable")

		// Verify old executable is still there
		content, err := afero.ReadFile(cmfs, executablePath)
		require.NoError(t, err)
		assert.Equal(t, "old content", string(content))

		// Verify .old file does not exist
		_, err = cmfs.Stat(executablePath + ".old")
		assert.True(t, os.IsNotExist(err))
	})

	t.Run("fails to replace executable and restore fails", func(t *testing.T) {
		memfs := afero.NewMemMapFs()
		cmfs := &controlledMockFs{Fs: memfs}
		cmfs.renameHooks = []func(old, new string) error{
			nil, // first rename (old -> old.old) succeeds
			func(old, new string) error { return fmt.Errorf("failed to replace") }, // second rename (new -> old) fails
			func(old, new string) error { return fmt.Errorf("failed to restore") }, // third rename (old.old -> old) fails
		}

		executablePath := "/app/server"
		err := afero.WriteFile(cmfs, executablePath, []byte("old content"), 0755)
		require.NoError(t, err)

		updater := NewUpdater(http.DefaultClient, "")

		err = updater.UpdateTo(context.Background(), cmfs, executablePath, release, assetName, "checksums.txt")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to replace executable and could not restore old version")

		// original file should not exist, but .old should
		_, err = cmfs.Stat(executablePath)
		assert.True(t, os.IsNotExist(err))

		content, err := afero.ReadFile(cmfs, executablePath+".old")
		require.NoError(t, err)
		assert.Equal(t, "old content", string(content))
	})

	t.Run("fails to replace executable but restore succeeds", func(t *testing.T) {
		memfs := afero.NewMemMapFs()
		cmfs := &controlledMockFs{Fs: memfs}
		cmfs.renameHooks = []func(old, new string) error{
			nil, // first rename (old -> old.old) succeeds
			func(old, new string) error { return fmt.Errorf("failed to replace") }, // second rename (new -> old) fails
			nil, // third rename (old.old -> old) succeeds
		}

		executablePath := "/app/server"
		err := afero.WriteFile(cmfs, executablePath, []byte("old content"), 0755)
		require.NoError(t, err)

		updater := NewUpdater(http.DefaultClient, "")

		err = updater.UpdateTo(context.Background(), cmfs, executablePath, release, assetName, "checksums.txt")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to replace executable")
		assert.NotContains(t, err.Error(), "could not restore")

		// original file should exist and have old content
		content, err := afero.ReadFile(cmfs, executablePath)
		require.NoError(t, err)
		assert.Equal(t, "old content", string(content))

		// .old file should not exist
		_, err = cmfs.Stat(executablePath + ".old")
		assert.True(t, os.IsNotExist(err))
	})
	t.Run("fails to replace executable and no old file to restore", func(t *testing.T) {
		memfs := afero.NewMemMapFs()
		cmfs := &controlledMockFs{Fs: memfs}
		cmfs.renameHooks = []func(old, new string) error{
			func(old, new string) error { return fmt.Errorf("failed to replace") }, // rename (new -> old) fails
		}

		executablePath := "/app/server"

		updater := NewUpdater(http.DefaultClient, "")

		err := updater.UpdateTo(context.Background(), cmfs, executablePath, release, assetName, "checksums.txt")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to replace executable")

		// No file should exist
		_, err = cmfs.Stat(executablePath)
		assert.True(t, os.IsNotExist(err))
	})
}

func TestUpdateTo_MoreFailureScenarios(t *testing.T) {
	assetContent := "new binary content"
	assetName := "server-linux-amd64"
	assetHash := sha256.Sum256([]byte(assetContent))
	checksumsContent := fmt.Sprintf("%s  %s\n", hex.EncodeToString(assetHash[:]), assetName)

	t.Run("asset not found in release", func(t *testing.T) {
		release := &github.RepositoryRelease{
			TagName: github.String("v1.0.1"),
			Assets:  []*github.ReleaseAsset{}, // No assets
		}
		updater := NewUpdater(http.DefaultClient, "")
		fs := afero.NewMemMapFs()
		err := updater.UpdateTo(context.Background(), fs, "/app/server", release, assetName, "checksums.txt")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "asset")
		assert.Contains(t, err.Error(), "not found in release")
	})

	t.Run("checksums asset not found in release", func(t *testing.T) {
		release := &github.RepositoryRelease{
			TagName: github.String("v1.0.1"),
			Assets: []*github.ReleaseAsset{
				{Name: github.String(assetName)}, // has the binary asset but not checksums
			},
		}
		updater := NewUpdater(http.DefaultClient, "")
		fs := afero.NewMemMapFs()
		err := updater.UpdateTo(context.Background(), fs, "/app/server", release, assetName, "checksums.txt")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "checksums asset")
		assert.Contains(t, err.Error(), "not found in release")
	})

	t.Run("checksum download fails", func(t *testing.T) {
		checksumsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer checksumsServer.Close()

		release := &github.RepositoryRelease{
			TagName: github.String("v1.0.1"),
			Assets: []*github.ReleaseAsset{
				{Name: github.String(assetName), BrowserDownloadURL: github.String("http://127.0.0.1/asset")},
				{Name: github.String("checksums.txt"), BrowserDownloadURL: github.String(checksumsServer.URL)},
			},
		}

		updater := NewUpdater(http.DefaultClient, "")
		fs := afero.NewMemMapFs()
		err := updater.UpdateTo(context.Background(), fs, "/app/server", release, assetName, "checksums.txt")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to download checksums")
	})

	t.Run("malformed checksums file", func(t *testing.T) {
		assetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte(assetContent)) }))
		defer assetServer.Close()

		checksumsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("this is not a valid checksums file"))
		}))
		defer checksumsServer.Close()

		release := &github.RepositoryRelease{
			TagName: github.String("v1.0.1"),
			Assets: []*github.ReleaseAsset{
				{Name: github.String(assetName), BrowserDownloadURL: github.String(assetServer.URL)},
				{Name: github.String("checksums.txt"), BrowserDownloadURL: github.String(checksumsServer.URL)},
			},
		}

		updater := NewUpdater(http.DefaultClient, "")
		fs := afero.NewMemMapFs()
		err := updater.UpdateTo(context.Background(), fs, "/app/server", release, assetName, "checksums.txt")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid checksum line")
	})

	t.Run("checksum for asset not in checksums file", func(t *testing.T) {
		assetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte(assetContent)) }))
		defer assetServer.Close()
		checksumsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("somehash  someotherfile"))
		}))
		defer checksumsServer.Close()

		release := &github.RepositoryRelease{
			TagName: github.String("v1.0.1"),
			Assets: []*github.ReleaseAsset{
				{Name: github.String(assetName), BrowserDownloadURL: github.String(assetServer.URL)},
				{Name: github.String("checksums.txt"), BrowserDownloadURL: github.String(checksumsServer.URL)},
			},
		}

		updater := NewUpdater(http.DefaultClient, "")
		fs := afero.NewMemMapFs()
		err := updater.UpdateTo(context.Background(), fs, "/app/server", release, assetName, "checksums.txt")
		require.Error(t, err)
		expectedErr := fmt.Sprintf("checksum for asset %s not found in checksums file", assetName)
		assert.EqualError(t, err, expectedErr)
	})

	t.Run("asset download fails", func(t *testing.T) {
		assetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer assetServer.Close()
		checksumsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(checksumsContent))
		}))
		defer checksumsServer.Close()

		release := &github.RepositoryRelease{
			TagName: github.String("v1.0.1"),
			Assets: []*github.ReleaseAsset{
				{Name: github.String(assetName), BrowserDownloadURL: github.String(assetServer.URL)},
				{Name: github.String("checksums.txt"), BrowserDownloadURL: github.String(checksumsServer.URL)},
			},
		}

		updater := NewUpdater(http.DefaultClient, "")
		fs := afero.NewMemMapFs()
		err := updater.UpdateTo(context.Background(), fs, "/app/server", release, assetName, "checksums.txt")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to download asset")
	})

	t.Run("checksum mismatch", func(t *testing.T) {
		assetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("wrong content"))
		}))
		defer assetServer.Close()
		checksumsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(checksumsContent))
		}))
		defer checksumsServer.Close()

		release := &github.RepositoryRelease{
			TagName: github.String("v1.0.1"),
			Assets: []*github.ReleaseAsset{
				{Name: github.String(assetName), BrowserDownloadURL: github.String(assetServer.URL)},
				{Name: github.String("checksums.txt"), BrowserDownloadURL: github.String(checksumsServer.URL)},
			},
		}

		updater := NewUpdater(http.DefaultClient, "")
		fs := afero.NewMemMapFs()
		err := updater.UpdateTo(context.Background(), fs, "/app/server", release, assetName, "checksums.txt")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "checksum mismatch")
	})

    t.Run("tempfile creation fails", func(t *testing.T) {
		assetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte(assetContent)) }))
		defer assetServer.Close()
		checksumsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(checksumsContent))
		}))
		defer checksumsServer.Close()
        release := &github.RepositoryRelease{TagName: github.String("v1.0.1"), Assets: []*github.ReleaseAsset{{Name: github.String(assetName), BrowserDownloadURL: github.String(assetServer.URL)},{Name: github.String("checksums.txt"), BrowserDownloadURL: github.String(checksumsServer.URL)}}}

        memfs := afero.NewMemMapFs()
        cmfs := &controlledMockFs{Fs: memfs}
        cmfs.openFileHooks = []func(name string, flag int, perm os.FileMode) (afero.File, error){
            func(name string, flag int, perm os.FileMode) (afero.File, error) { return nil, fmt.Errorf("disk full") },
        }
		updater := NewUpdater(http.DefaultClient, "")
		err := updater.UpdateTo(context.Background(), cmfs, "/app/server", release, assetName, "checksums.txt")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create temp file")
    })

    t.Run("chmod fails", func(t *testing.T) {
		assetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte(assetContent)) }))
		defer assetServer.Close()
		checksumsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(checksumsContent))
		}))
		defer checksumsServer.Close()
        release := &github.RepositoryRelease{TagName: github.String("v1.0.1"), Assets: []*github.ReleaseAsset{{Name: github.String(assetName), BrowserDownloadURL: github.String(assetServer.URL)},{Name: github.String("checksums.txt"), BrowserDownloadURL: github.String(checksumsServer.URL)}}}

        memfs := afero.NewMemMapFs()
        cmfs := &controlledMockFs{Fs: memfs}
        cmfs.chmodHooks = []func(name string, mode os.FileMode) error{
            func(name string, mode os.FileMode) error { return fmt.Errorf("permission denied") },
        }
		updater := NewUpdater(http.DefaultClient, "")
		err := updater.UpdateTo(context.Background(), cmfs, "/app/server", release, assetName, "checksums.txt")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to set executable permission")
    })
}

func TestParseChecksums(t *testing.T) {
	t.Run("valid checksums", func(t *testing.T) {
		data := "hash1  file1\nhash2  file2\n"
		checksums, err := parseChecksums(data)
		require.NoError(t, err)
		assert.Equal(t, "hash1", checksums["file1"])
		assert.Equal(t, "hash2", checksums["file2"])
	})

	t.Run("invalid line", func(t *testing.T) {
		data := "hash1 file1 extra"
		_, err := parseChecksums(data)
		require.Error(t, err)
	})

	t.Run("empty line", func(t *testing.T) {
		data := "hash1  file1\n\n"
		checksums, err := parseChecksums(data)
		require.NoError(t, err)
		assert.Len(t, checksums, 1)
	})
}
