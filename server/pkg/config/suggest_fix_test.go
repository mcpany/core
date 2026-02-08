package config

import (
	"context"
	"strings"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestSuggestFix_CommonAlias_Services(t *testing.T) {
	fs := afero.NewMemMapFs()
	// Config using "services" which is a common mistake for "upstream_services"
	badConfig := `
version: "1"
services:
  - name: "my-service"
`
	_ = afero.WriteFile(fs, "services_alias.yaml", []byte(badConfig), 0644)

	store := NewFileStore(fs, []string{"services_alias.yaml"})
	_, err := store.Load(context.Background())

	assert.Error(t, err)
	// Verify the specific helpful message
	assert.True(t, strings.Contains(err.Error(), "Did you mean \"upstream_services\"? \"services\" is not a valid top-level key."), "Error message should contain specific hint for 'services'")
}

func TestSuggestFix_Recursion_Excluded(t *testing.T) {
	// We want to ensure that fields from irrelevant messages (like Collection which has 'services')
	// are NOT suggested when we are at the root level.
	// We can't easily test 'Collection' exclusion specifically without mocking, but we can test
	// that we DO get suggestions for common types.

	fs := afero.NewMemMapFs()
	// Typo "http_servic" -> "http_service" (in UpstreamServiceConfig)
	badConfig := `
upstream_services:
  - name: "test"
    http_servic:
      address: "http://127.0.0.1"
`
	_ = afero.WriteFile(fs, "typo_http.yaml", []byte(badConfig), 0644)

	store := NewFileStore(fs, []string{"typo_http.yaml"})
	_, err := store.Load(context.Background())

	assert.Error(t, err)
	// "http_servic" is close to "http_service" which is in UpstreamServiceConfig (one of the allowed common messages)
	assert.True(t, strings.Contains(err.Error(), "Did you mean \"http_service\"?"), "Should suggest http_service")
}

func TestSuggestFix_DeeplyNested_Included(t *testing.T) {
	// "address" is in HttpUpstreamService, which is in our allowed list.
	fs := afero.NewMemMapFs()
	badConfig := `
upstream_services:
  - name: "test"
    http_service:
      addres: "http://127.0.0.1"
`
	// "addres" -> "address"
	_ = afero.WriteFile(fs, "typo_address.yaml", []byte(badConfig), 0644)

	store := NewFileStore(fs, []string{"typo_address.yaml"})
	_, err := store.Load(context.Background())

	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "Did you mean \"address\"?"), "Should suggest address from HttpUpstreamService")
}
