/*
 * Copyright 2025 Author(s) of MCP-XY
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package framework

import (
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type GeminiCLI struct {
	t *testing.T
}

func NewGeminiCLI(t *testing.T) *GeminiCLI {
	return &GeminiCLI{t: t}
}

func (g *GeminiCLI) Install() {
	g.t.Helper()
	cmd := exec.Command("npm", "install", "-g", "@google/gemini-cli")
	err := cmd.Run()
	require.NoError(g.t, err, "failed to install gemini-cli")
}

func (g *GeminiCLI) AddMCP(name, endpoint string) {
	g.t.Helper()
	cmd := exec.Command("gemini", "mcp", "add", "--transport", "http", name, endpoint)
	err := cmd.Run()
	require.NoError(g.t, err, "failed to configure gemini-cli")
}

func (g *GeminiCLI) RemoveMCP(name string) {
	g.t.Helper()
	cmd := exec.Command("gemini", "mcp", "remove", name)
	err := cmd.Run()
	if err != nil {
		g.t.Logf("failed to remove mcp server '%s': %v", name, err)
	}
}

func (g *GeminiCLI) Run(apiKey, model, prompt string) (string, error) {
	g.t.Helper()
	os.Setenv("GEMINI_API_KEY", apiKey)
	defer os.Unsetenv("GEMINI_API_KEY")

	var outputBuffer strings.Builder
	cmd := exec.Command("gemini", "-m", model, "-p", prompt)
	cmd.Stdout = &outputBuffer
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	return outputBuffer.String(), err
}
