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
	"strings"

	"github.com/google/go-github/v39/github"
	"github.com/spf13/afero"
)

// Updater represents the public Updater entity.
//
// Summary: Defines the structured data model representing a .
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
type Updater struct {
	client     *github.Client
	httpClient *http.Client
}

// NewUpdater serves as a public interface for interacting with NewUpdater.
//
// Summary: Constructs and returns an initialized updater ready for consumption.
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
func NewUpdater(httpClient *http.Client, githubAPIURL string) *Updater {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	client := github.NewClient(httpClient)
	if githubAPIURL != "" {
		baseURL, err := url.Parse(githubAPIURL)
		if err == nil {
			if !strings.HasSuffix(baseURL.Path, "/") {
				baseURL.Path += "/"
			}
			client.BaseURL = baseURL
		}
	}
	return &Updater{client: client, httpClient: httpClient}
}

// CheckForUpdate serves as a public interface for interacting with CheckForUpdate.
//
// Summary: Check the for update appropriately based on current system conditions.
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

// UpdateTo serves as a public interface for interacting with UpdateTo.
//
// Summary: Update the to appropriately based on current system conditions.
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
	if resp.StatusCode >= 400 {
		return fmt.Errorf("failed to download checksums: status code %d", resp.StatusCode)
	}
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
	if resp.StatusCode >= 400 {
		return fmt.Errorf("failed to download asset: status code %d", resp.StatusCode)
	}

	// Create a temporary file to save the downloaded asset
	tmpFile, err := afero.TempFile(fs, "", "mcpany-update-")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}

	// Write the downloaded asset to the temp file and calculate the checksum
	hasher := sha256.New()
	if _, err := io.Copy(io.MultiWriter(tmpFile, hasher), resp.Body); err != nil {
		_ = tmpFile.Close() // Close the file on error
		return fmt.Errorf("failed to write to temp file: %w", err)
	}
	actualChecksum := hex.EncodeToString(hasher.Sum(nil))

	// Close the file before renaming it
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	// Verify the checksum
	if actualChecksum != expectedChecksum {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedChecksum, actualChecksum)
	}

	if err := fs.Chmod(tmpFile.Name(), 0o755); err != nil {
		return fmt.Errorf("failed to set executable permission: %w", err)
	}

	// On Windows, we cannot replace a running executable.
	// The workaround is to rename the old executable, move the new one,
	// and the old one can be cleaned up later.
	oldPath := executablePath + ".old"
	if _, err := fs.Stat(executablePath); err == nil {
		if err := fs.Rename(executablePath, oldPath); err != nil {
			return fmt.Errorf("failed to rename old executable: %w", err)
		}
	}

	if err := fs.Rename(tmpFile.Name(), executablePath); err != nil {
		// If the rename fails, try to restore the old executable.
		if _, err := fs.Stat(oldPath); err == nil {
			if err := fs.Rename(oldPath, executablePath); err != nil {
				return fmt.Errorf("failed to replace executable and could not restore old version: %w", err)
			}
		}
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
