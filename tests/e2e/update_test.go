// Copyright 2025 Author(s) of MCP Any
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

//go:build e2e

package e2e

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"runtime"
	"testing"

	"github.com/google/go-github/v39/github"
	"github.com/stretchr/testify/assert"
)

func TestUpdateCommand(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping update test on Windows due to file locking issues with running executables.")
	}

	// Prepare the mock servers
	assetName := fmt.Sprintf("server-%s-%s", runtime.GOOS, runtime.GOARCH)
	assetContent := "this is a fake binary"
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

	githubServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
		json.NewEncoder(w).Encode(release)
	}))
	defer githubServer.Close()

	// Create a dummy file to be updated
	tempDir, err := os.MkdirTemp("", "mcpany-e2e-update-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	dummyExecutablePath := tempDir + "/dummy-server"
	err = os.WriteFile(dummyExecutablePath, []byte("old content"), 0600)
	assert.NoError(t, err)

	// Run the update command against the dummy file
	updateCmd := exec.Command("go", "run", "../../cmd/server", "update", "--path", dummyExecutablePath)
	updateCmd.Env = append(os.Environ(), "GITHUB_API_URL="+githubServer.URL)
	updateOutput, err := updateCmd.CombinedOutput()
	assert.NoError(t, err, string(updateOutput))

	// Verify the output and the new executable
	assert.Contains(t, string(updateOutput), "A new version is available: v1.0.1. Updating...")
	assert.Contains(t, string(updateOutput), "Update successful.")

	updatedContent, err := os.ReadFile(dummyExecutablePath)
	assert.NoError(t, err)
	assert.Equal(t, assetContent, string(updatedContent))
}
