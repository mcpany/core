commit 40f9d4805b2f5edf7e672789c4fece4dd8a3384f
Author: ql-owo-lp <678729+ql-owo-lp@users.noreply.github.com>
Date:   Fri Mar 27 00:23:31 2026 +0000

    chore: remove restricted github files

diff --git a/server/cmd/mcpctl/validate_test.go b/server/cmd/mcpctl/validate_test.go
index b6e223375..1d59d7983 100644
--- a/server/cmd/mcpctl/validate_test.go
+++ b/server/cmd/mcpctl/validate_test.go
@@ -1,50 +1,114 @@
-commit 034ba7b541aed499b065db5d8b4bc2c9dc8234bc
-Author: ql-owo-lp <678729+ql-owo-lp@users.noreply.github.com>
-Date:   Thu Mar 26 02:03:04 2026 +0000
+// Copyright 2025 Author(s) of MCP Any
+// SPDX-License-Identifier: Apache-2.0

-    fix: Truth Reconciliation Audit and Test Fixes
+package main

-diff --git a/server/cmd/mcpctl/validate_test.go b/server/cmd/mcpctl/validate_test.go
-index 1d59d7983..4b743d85e 100644
---- a/server/cmd/mcpctl/validate_test.go
-+++ b/server/cmd/mcpctl/validate_test.go
-@@ -9,7 +9,6 @@ import (
+import (
+	"bytes"
+	"os"
	"path/filepath"
	"testing"

--	"github.com/spf13/viper"
+	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
- )
-@@ -38,7 +37,6 @@ upstream_services:
+)
+
+func TestValidateCmd(t *testing.T) {
+	tempDir := t.TempDir()
+
+	// 1. Valid Configuration
+	validConfig := `
+global_settings:
+  mcp_listen_address: "localhost:50050"
+upstream_services:
+  - name: "my-service"
+    http_service:
+      address: "http://example.com"
+      tools:
+        - name: "my-tool"
+          call_id: "my-call"
+      calls:
+        my-call:
+          endpoint_path: "/api"
+          method: "HTTP_METHOD_GET"
+`
+	validConfigPath := filepath.Join(tempDir, "valid_config.yaml")
+	err := os.WriteFile(validConfigPath, []byte(validConfig), 0644)
	require.NoError(t, err)

	t.Run("Valid Configuration", func(t *testing.T) {
--		viper.Reset()
+		viper.Reset()
		cmd := newRootCmd()
		b := bytes.NewBufferString("")
		cmd.SetOut(b)
-@@ -70,14 +68,13 @@ upstream_services:
+		cmd.SetArgs([]string{"validate", "--config-path", validConfigPath})
+		err := cmd.Execute()
+		assert.NoError(t, err)
+		assert.Contains(t, b.String(), "Configuration is valid.")
+	})
+
+	// 2. Invalid YAML Syntax
+	invalidYAML := `
+global_settings:
+  mcp_listen_address: "localhost:50050"
+upstream_services:
+  - name: "my-service"
+    http_service:
+      address: "http://example.com"
+      tools:
+        - name: "my-tool"
+          call_id: "my-call"
+      calls:
+        my-call:
+          endpoint_path: "/api"
+          method: "HTTP_METHOD_GET"
+    invalid_indentation
+`
+	invalidYAMLPath := filepath.Join(tempDir, "invalid_yaml.yaml")
+	err = os.WriteFile(invalidYAMLPath, []byte(invalidYAML), 0644)
	require.NoError(t, err)

	t.Run("Invalid YAML Syntax", func(t *testing.T) {
--		viper.Reset()
+		viper.Reset()
		cmd := newRootCmd()
		b := bytes.NewBufferString("")
		cmd.SetOut(b)
		cmd.SetArgs([]string{"validate", "--config-path", invalidYAMLPath})
		err := cmd.Execute()
		assert.Error(t, err)
--		assert.Contains(t, err.Error(), "failed to unmarshal")
-+		assert.Contains(t, err.Error(), "failed to unmarshal config")
+		assert.Contains(t, err.Error(), "failed to unmarshal")
	})

	// 3. Validation Error (Invalid HTTP address)
-@@ -101,7 +98,6 @@ upstream_services:
+	invalidConfig := `
+global_settings:
+  mcp_listen_address: "localhost:50050"
+upstream_services:
+  - name: "my-service"
+    http_service:
+      address: "::invalid::"
+      tools:
+        - name: "my-tool"
+          call_id: "my-call"
+      calls:
+        my-call:
+          endpoint_path: "/api"
+          method: "HTTP_METHOD_GET"
+`
+	invalidConfigPath := filepath.Join(tempDir, "invalid_config.yaml")
+	err = os.WriteFile(invalidConfigPath, []byte(invalidConfig), 0644)
	require.NoError(t, err)

	t.Run("Validation Error", func(t *testing.T) {
--		viper.Reset()
+		viper.Reset()
		cmd := newRootCmd()
		b := bytes.NewBufferString("")
		cmd.SetOut(b)
+		cmd.SetArgs([]string{"validate", "--config-path", invalidConfigPath})
+		err := cmd.Execute()
+		assert.Error(t, err)
+		assert.Contains(t, err.Error(), "Configuration Validation Failed")
+		assert.Contains(t, err.Error(), "invalid http address")
+	})
+}
