/*
 * Copyright 2025 Author(s) of MCP Any
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

package config

import (
	"os"
	"os/exec"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestBindRootFlags_Error(t *testing.T) {
	if os.Getenv("GO_TEST_SUBPROCESS") == "1" {
		cmd := &cobra.Command{}
		cmd.PersistentFlags().String("mcp-listen-address", "", "")
		BindRootFlags(cmd)
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestBindRootFlags_Error")
	cmd.Env = append(os.Environ(), "GO_TEST_SUBPROCESS=1")
	err := cmd.Run()

	assert.Error(t, err, "expected command to fail")
	if e, ok := err.(*exec.ExitError); ok {
		assert.False(t, e.Success(), "expected command to exit with non-zero status")
	}
}

func TestBindServerFlags_Error(t *testing.T) {
	if os.Getenv("GO_TEST_SUBPROCESS") == "1" {
		cmd := &cobra.Command{}
		cmd.Flags().String("grpc-port", "", "")
		BindServerFlags(cmd)
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestBindServerFlags_Error")
	cmd.Env = append(os.Environ(), "GO_TEST_SUBPROCESS=1")
	err := cmd.Run()

	assert.Error(t, err, "expected command to fail")
	if e, ok := err.(*exec.ExitError); ok {
		assert.False(t, e.Success(), "expected command to exit with non-zero status")
	}
}
