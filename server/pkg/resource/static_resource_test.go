// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package resource

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func strPtr(s string) *string { return &s }
func int64Ptr(i int64) *int64 { return &i }

func TestStaticResource(t *testing.T) {
	t.Setenv("MCPANY_ALLOW_LOOPBACK_RESOURCES", "true")

	// Setup test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/test.txt" {
			w.Header().Set("Content-Type", "text/plain")
			_, _ = w.Write([]byte("hello world"))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	def := &configv1.ResourceDefinition{
		Uri:         strPtr(server.URL + "/test.txt"),
		Name:        strPtr("Test Resource"),
		Description: strPtr("A test resource"),
		MimeType:    strPtr("text/plain"),
		Size:        int64Ptr(11),
	}

	serviceID := "test-service"
	r := NewStaticResource(def, serviceID)

	t.Run("Metadata", func(t *testing.T) {
		assert.Equal(t, serviceID, r.Service())
		res := r.Resource()
		assert.Equal(t, def.GetUri(), res.URI)
		assert.Equal(t, def.GetName(), res.Name)
		assert.Equal(t, def.GetDescription(), res.Description)
		assert.Equal(t, def.GetMimeType(), res.MIMEType)
		assert.Equal(t, def.GetSize(), res.Size)
	})

	t.Run("Read", func(t *testing.T) {
		res, err := r.Read(context.Background())
		require.NoError(t, err)
		require.Len(t, res.Contents, 1)
		content := res.Contents[0]
		assert.Equal(t, def.GetUri(), content.URI)
		assert.Equal(t, []byte("hello world"), content.Blob)
		assert.Equal(t, def.GetMimeType(), content.MIMEType)
	})

	t.Run("ReadError", func(t *testing.T) {
		badDef := &configv1.ResourceDefinition{
			Uri: strPtr(server.URL + "/404"),
		}
		badR := NewStaticResource(badDef, serviceID)
		_, err := badR.Read(context.Background())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "status: 404")
	})

	t.Run("ReadNetworkError", func(t *testing.T) {
		badDef := &configv1.ResourceDefinition{
			Uri: strPtr("http://localhost:0"), // Invalid port
		}
		badR := NewStaticResource(badDef, serviceID)
		_, err := badR.Read(context.Background())
		assert.Error(t, err)
	})

	t.Run("Subscribe", func(t *testing.T) {
		err := r.Subscribe(context.Background())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not yet implemented")
	})
}
