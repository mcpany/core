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

package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCloneGitRepository(t *testing.T) {
	// Clone a known public repository.
	tempDir, err := CloneGitRepository("https://github.com/git-fixtures/basic.git")
	assert.NoError(t, err, "CloneGitRepository should not return an error")
	defer os.RemoveAll(tempDir)

	// Verify that the repository was cloned successfully by checking for the presence of a known file.
	_, err = os.Stat(filepath.Join(tempDir, ".git"))
	assert.NoError(t, err, ".git should exist in the cloned repository")
}
