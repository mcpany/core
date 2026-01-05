// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package update

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
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
	renameHooks []func(old, new string) error
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

		updater := NewUpdater(http.DefaultClient)

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

		updater := NewUpdater(http.DefaultClient)

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

		updater := NewUpdater(http.DefaultClient)

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
}
