commit 034ba7b541aed499b065db5d8b4bc2c9dc8234bc
Author: ql-owo-lp <678729+ql-owo-lp@users.noreply.github.com>
Date:   Thu Mar 26 02:03:04 2026 +0000

    fix: Truth Reconciliation Audit and Test Fixes

diff --git a/server/cmd/mcpctl/validate_test.go b/server/cmd/mcpctl/validate_test.go
index 1d59d7983..4b743d85e 100644
--- a/server/cmd/mcpctl/validate_test.go
+++ b/server/cmd/mcpctl/validate_test.go
@@ -9,7 +9,6 @@ import (
	"path/filepath"
	"testing"

-	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
 )
@@ -38,7 +37,6 @@ upstream_services:
	require.NoError(t, err)

	t.Run("Valid Configuration", func(t *testing.T) {
-		viper.Reset()
		cmd := newRootCmd()
		b := bytes.NewBufferString("")
		cmd.SetOut(b)
@@ -70,14 +68,13 @@ upstream_services:
	require.NoError(t, err)

	t.Run("Invalid YAML Syntax", func(t *testing.T) {
-		viper.Reset()
		cmd := newRootCmd()
		b := bytes.NewBufferString("")
		cmd.SetOut(b)
		cmd.SetArgs([]string{"validate", "--config-path", invalidYAMLPath})
		err := cmd.Execute()
		assert.Error(t, err)
-		assert.Contains(t, err.Error(), "failed to unmarshal")
+		assert.Contains(t, err.Error(), "failed to unmarshal config")
	})

	// 3. Validation Error (Invalid HTTP address)
@@ -101,7 +98,6 @@ upstream_services:
	require.NoError(t, err)

	t.Run("Validation Error", func(t *testing.T) {
-		viper.Reset()
		cmd := newRootCmd()
		b := bytes.NewBufferString("")
		cmd.SetOut(b)
