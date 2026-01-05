// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package update provides functionality for self-updating the application.
package update

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/google/go-github/v39/github"
	"github.com/spf13/afero"
)

// Updater handles the self-update process.
type Updater struct {
	client     *github.Client
	httpClient *http.Client
}

// NewUpdater creates a new Updater.
func NewUpdater(httpClient *http.Client) *Updater {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	client := github.NewClient(httpClient)
	if os.Getenv("GITHUB_API_URL") != "" {
		baseURL, err := url.Parse(os.Getenv("GITHUB_API_URL"))
		if err == nil {
			if !strings.HasSuffix(baseURL.Path, "/") {
				baseURL.Path += "/"
			}
			client.BaseURL = baseURL
		}
	}
	return &Updater{client: client, httpClient: httpClient}
}

// CheckForUpdate checks for a new release on GitHub.
func (u *Updater) CheckForUpdate(ctx context.Context, owner, repo, currentVersion string) (*github.RepositoryRelease, bool, error) {
	release, _, err := u.client.Repositories.GetLatestRelease(ctx, owner, repo)
	if err != nil {
		return nil, false, fmt.Errorf("failed to get latest release: %w", err)
	}

	if release.GetTagName() == currentVersion {
		return nil, false, nil
	}

	return release, true, nil
}

// UpdateTo downloads the new release, verifies its checksum, and replaces the current executable.
func (u *Updater) UpdateTo(ctx context.Context, fs afero.Fs, executablePath string, release *github.RepositoryRelease, assetName, checksumsAssetName string) error {
	var asset *github.ReleaseAsset
	for _, a := range release.Assets {
		if a.GetName() == assetName {
			asset = a
			break
		}
	}
	if asset == nil {
		return fmt.Errorf("asset %s not found in release %s", assetName, release.GetTagName())
	}

	var checksumsAsset *github.ReleaseAsset
	for _, a := range release.Assets {
		if a.GetName() == checksumsAssetName {
			checksumsAsset = a
			break
		}
	}
	if checksumsAsset == nil {
		return fmt.Errorf("checksums asset %s not found in release %s", checksumsAssetName, release.GetTagName())
	}

	// Download the checksums file
	req, err := http.NewRequestWithContext(ctx, "GET", checksumsAsset.GetBrowserDownloadURL(), nil)
	if err != nil {
		return fmt.Errorf("failed to create request for checksums: %w", err)
	}
	resp, err := u.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download checksums: %w", err)
	}
	defer func(r *http.Response) { _ = r.Body.Close() }(resp)
	checksumsData, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read checksums data: %w", err)
	}
	checksums, err := parseChecksums(string(checksumsData))
	if err != nil {
		return fmt.Errorf("failed to parse checksums: %w", err)
	}
	expectedChecksum, ok := checksums[assetName]
	if !ok {
		return fmt.Errorf("checksum for asset %s not found in checksums file", assetName)
	}

	// Download the release asset
	req, err = http.NewRequestWithContext(ctx, "GET", asset.GetBrowserDownloadURL(), nil)
	if err != nil {
		return fmt.Errorf("failed to create request for asset: %w", err)
	}
	resp, err = u.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download asset: %w", err)
	}
	defer func(r *http.Response) { _ = r.Body.Close() }(resp)

	// Create a temporary file to save the downloaded asset
	tmpFile, err := afero.TempFile(fs, "", "mcpany-update-")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer func() { _ = tmpFile.Close() }()

	// Write the downloaded asset to the temp file and calculate the checksum
	hasher := sha256.New()
	if _, err := io.Copy(io.MultiWriter(tmpFile, hasher), resp.Body); err != nil {
		return fmt.Errorf("failed to write to temp file: %w", err)
	}
	actualChecksum := hex.EncodeToString(hasher.Sum(nil))

	// Verify the checksum
	if actualChecksum != expectedChecksum {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedChecksum, actualChecksum)
	}

	if err := fs.Chmod(tmpFile.Name(), 0o755); err != nil {
		return fmt.Errorf("failed to set executable permission: %w", err)
	}

	if err := fs.Rename(tmpFile.Name(), executablePath); err != nil {
		return fmt.Errorf("failed to replace executable: %w", err)
	}

	return nil
}

// parseChecksums parses the checksums file and returns a map of filenames to checksums.
func parseChecksums(data string) (map[string]string, error) {
	checksums := make(map[string]string)
	lines := strings.Split(data, "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid checksum line: %s", line)
		}
		checksums[parts[1]] = parts[0]
	}
	return checksums, nil
}
