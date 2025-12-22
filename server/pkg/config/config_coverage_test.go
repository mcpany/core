package config

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestBindRootFlags(t *testing.T) {
	cmd := &cobra.Command{}
	BindRootFlags(cmd)

	// Verify flags are present
	assert.NotNil(t, cmd.PersistentFlags().Lookup("config-path"))
	assert.NotNil(t, cmd.PersistentFlags().Lookup("log-level"))
	assert.NotNil(t, cmd.PersistentFlags().Lookup("logfile"))
}

func TestBindServerFlags(t *testing.T) {
	cmd := &cobra.Command{}
	BindServerFlags(cmd)

	// Verify flags are present
	assert.NotNil(t, cmd.Flags().Lookup("grpc-port"))
	assert.NotNil(t, cmd.Flags().Lookup("stdio"))
}
