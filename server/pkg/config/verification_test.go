// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
	"testing"

	"github.com/spf13/afero"
    "github.com/stretchr/testify/require"
)

func TestVerificationConfig(t *testing.T) {
	fs := afero.NewOsFs()
	store := NewFileStore(fs, []string{"../../../verification_config.yaml"})

	ctx := context.Background()
	config, err := store.Load(ctx)

	require.NoError(t, err)
    require.NotNil(t, config)
}
