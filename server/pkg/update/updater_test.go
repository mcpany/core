package update

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"runtime"
	"testing"

	"github.com/google/go-github/v39/github"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdater(t *testing.T) {
	const executablePath = "/usr/local/bin/server"
	assetName := fmt.Sprintf("server-%s-%s", runtime.GOOS, runtime.GOARCH)
	assetContent := "dummy content"
	assetHash := sha256.Sum256([]byte(assetContent))
	checksumsContent := fmt.Sprintf("%s  %s\n", hex.EncodeToString(assetHash[:]), assetName)

	assetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(assetContent))
	}))
	defer assetServer.Close()

	checksumsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(checksumsContent))
	}))
	defer checksumsServer.Close()

	t.Run("UpdateTo", func(t *testing.T) {
		t.Run("success", func(t *testing.T) {
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
			fs := afero.NewMemMapFs()
			// executablePath defined in outer scope
			err := afero.WriteFile(fs, executablePath, []byte("old content"), 0o755)
			require.NoError(t, err)

			updater := NewUpdater(nil)
			err = updater.UpdateTo(fs, executablePath, release, assetName, "checksums.txt")
			assert.NoError(t, err)

			content, err := afero.ReadFile(fs, executablePath)
			assert.NoError(t, err)
			assert.Equal(t, assetContent, string(content))
		})

		t.Run("asset not found", func(t *testing.T) {
			release := &github.RepositoryRelease{
				TagName: github.String("v1.0.1"),
				Assets:  []*github.ReleaseAsset{},
			}
			fs := afero.NewMemMapFs()
			// executablePath defined in outer scope
			err := afero.WriteFile(fs, executablePath, []byte("old content"), 0o755)
			require.NoError(t, err)

			updater := NewUpdater(nil)
			err = updater.UpdateTo(fs, executablePath, release, assetName, "checksums.txt")
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "asset")
			assert.Contains(t, err.Error(), "not found")
		})

		t.Run("checksums not found", func(t *testing.T) {
			release := &github.RepositoryRelease{
				TagName: github.String("v1.0.1"),
				Assets: []*github.ReleaseAsset{
					{
						Name:               github.String(assetName),
						BrowserDownloadURL: github.String(assetServer.URL),
					},
				},
			}
			fs := afero.NewMemMapFs()
			// executablePath defined in outer scope
			err := afero.WriteFile(fs, executablePath, []byte("old content"), 0o755)
			require.NoError(t, err)

			updater := NewUpdater(nil)
			err = updater.UpdateTo(fs, executablePath, release, assetName, "checksums.txt")
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "checksums")
			assert.Contains(t, err.Error(), "not found")
		})

		t.Run("checksum mismatch", func(t *testing.T) {
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
			fs := afero.NewMemMapFs()
			// executablePath defined in outer scope
			err := afero.WriteFile(fs, executablePath, []byte("old content"), 0o755)
			require.NoError(t, err)

			updater := NewUpdater(nil)
			// create a new checksums server with a bad checksum
			badChecksumsContent := "badchecksum  " + assetName + "\n"
			badChecksumsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				_, _ = w.Write([]byte(badChecksumsContent))
			}))
			defer badChecksumsServer.Close()
			release.Assets[1].BrowserDownloadURL = github.String(badChecksumsServer.URL)

			err = updater.UpdateTo(fs, executablePath, release, assetName, "checksums.txt")
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "checksum mismatch")
		})

		t.Run("download asset fails", func(t *testing.T) {
			release := &github.RepositoryRelease{
				TagName: github.String("v1.0.1"),
				Assets: []*github.ReleaseAsset{
					{
						Name:               github.String(assetName),
						BrowserDownloadURL: github.String("badurl"),
					},
					{
						Name:               github.String("checksums.txt"),
						BrowserDownloadURL: github.String(checksumsServer.URL),
					},
				},
			}
			fs := afero.NewMemMapFs()
			// executablePath defined in outer scope
			err := afero.WriteFile(fs, executablePath, []byte("old content"), 0o755)
			require.NoError(t, err)

			updater := NewUpdater(nil)
			err = updater.UpdateTo(fs, executablePath, release, assetName, "checksums.txt")
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "failed to download asset")
		})

		t.Run("download checksums fails", func(t *testing.T) {
			release := &github.RepositoryRelease{
				TagName: github.String("v1.0.1"),
				Assets: []*github.ReleaseAsset{
					{
						Name:               github.String(assetName),
						BrowserDownloadURL: github.String(assetServer.URL),
					},
					{
						Name:               github.String("checksums.txt"),
						BrowserDownloadURL: github.String("badurl"),
					},
				},
			}
			fs := afero.NewMemMapFs()
			// executablePath defined in outer scope
			err := afero.WriteFile(fs, executablePath, []byte("old content"), 0o755)
			require.NoError(t, err)

			updater := NewUpdater(nil)
			err = updater.UpdateTo(fs, executablePath, release, assetName, "checksums.txt")
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "failed to download checksums")
		})

		t.Run("malformed checksums file", func(t *testing.T) {
			malformedChecksumsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				_, _ = w.Write([]byte(`{"latest_version": "v1.2.3"}`))
			}))
			defer malformedChecksumsServer.Close()
			release := &github.RepositoryRelease{
				TagName: github.String("v1.0.1"),
				Assets: []*github.ReleaseAsset{
					{
						Name:               github.String(assetName),
						BrowserDownloadURL: github.String(assetServer.URL),
					},
					{
						Name:               github.String("checksums.txt"),
						BrowserDownloadURL: github.String(malformedChecksumsServer.URL),
					},
				},
			}
			fs := afero.NewMemMapFs()
			// executablePath defined in outer scope
			err := afero.WriteFile(fs, executablePath, []byte("old content"), 0o755)
			require.NoError(t, err)

			updater := NewUpdater(nil)
			err = updater.UpdateTo(fs, executablePath, release, assetName, "checksums.txt")
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "checksum for asset")
		})
	})
	t.Run("CheckForUpdate", func(t *testing.T) {
		t.Run("new version available", func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				release := &github.RepositoryRelease{
					TagName: github.String("v1.0.1"),
				}
				_ = json.NewEncoder(w).Encode(release)
			}))
			defer server.Close()

			client := github.NewClient(nil)
			url, _ := url.Parse(server.URL + "/")
			client.BaseURL = url

			updater := &Updater{client: client}
			release, available, err := updater.CheckForUpdate(context.Background(), "owner", "repo", "v1.0.0")
			assert.NoError(t, err)
			assert.True(t, available)
			assert.NotNil(t, release)
			assert.Equal(t, "v1.0.1", release.GetTagName())
		})
		t.Run("no new version", func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				release := &github.RepositoryRelease{
					TagName: github.String("v1.0.0"),
				}
				_ = json.NewEncoder(w).Encode(release)
			}))
			defer server.Close()

			client := github.NewClient(nil)
			url, _ := url.Parse(server.URL + "/")
			client.BaseURL = url

			updater := &Updater{client: client}
			release, available, err := updater.CheckForUpdate(context.Background(), "owner", "repo", "v1.0.0")
			assert.NoError(t, err)
			assert.False(t, available)
			assert.Nil(t, release)
		})

		t.Run("api error", func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			}))
			defer server.Close()

			client := github.NewClient(nil)
			url, _ := url.Parse(server.URL + "/")
			client.BaseURL = url

			updater := &Updater{client: client}
			_, available, err := updater.CheckForUpdate(context.Background(), "owner", "repo", "v1.0.0")
			assert.Error(t, err)
			assert.False(t, available)
		})
	})
}
