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
			Uri: strPtr("http://127.0.0.1:0"), // Invalid port
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

	t.Run("ReadSizeLimit", func(t *testing.T) {
		// Server returns "hello world" (11 bytes).
		// Set limit to 5.
		limitDef := &configv1.ResourceDefinition{
			Uri:  strPtr(server.URL + "/test.txt"),
			Size: int64Ptr(5),
		}
		limitR := NewStaticResource(limitDef, serviceID)
		_, err := limitR.Read(context.Background())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "exceeds limit")
	})

	t.Run("ReadNewRequestError", func(t *testing.T) {
		// NewRequestWithContext fails on bad URI characters, e.g. control chars.
		badDef := &configv1.ResourceDefinition{
			Uri: strPtr("http://example.com/\x00"),
		}
		badR := NewStaticResource(badDef, serviceID)
		_, err := badR.Read(context.Background())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create request")
	})

	t.Run("ReadMimeTypeFallback", func(t *testing.T) {
		expectedMimeType := "application/json"
		fallbackServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", expectedMimeType)
			_, _ = w.Write([]byte(`{"key": "value"}`))
		}))
		defer fallbackServer.Close()

		fallbackDef := &configv1.ResourceDefinition{
			Uri:  strPtr(fallbackServer.URL),
			Name: strPtr("Test Resource"),
		}
		fallbackR := NewStaticResource(fallbackDef, serviceID)
		res, err := fallbackR.Read(context.Background())
		require.NoError(t, err)
		require.Len(t, res.Contents, 1)
		content := res.Contents[0]
		assert.Equal(t, expectedMimeType, content.MIMEType, "MIMEType should fall back to Content-Type header")
	})
}

func TestStaticResource_InlineContent(t *testing.T) {
	t.Run("TextContent", func(t *testing.T) {
		textContent := "Hello, Inline World!"
		uri := "internal://hello"

		def := &configv1.ResourceDefinition{
			Uri:      &uri,
			Name:     strPtr("Inline Resource"),
			MimeType: strPtr("text/plain"),
			ResourceType: &configv1.ResourceDefinition_Static{
				Static: &configv1.StaticResource{
					ContentType: &configv1.StaticResource_TextContent{
						TextContent: textContent,
					},
				},
			},
		}

		r := NewStaticResource(def, "test-service")

		res, err := r.Read(context.Background())

		require.NoError(t, err, "Read should not return error for inline content")
		require.NotNil(t, res)
		require.Len(t, res.Contents, 1)
		assert.Equal(t, textContent, string(res.Contents[0].Blob))
		assert.Equal(t, "text/plain", res.Contents[0].MIMEType)
	})

	t.Run("BinaryContent", func(t *testing.T) {
		binaryContent := []byte{0xDE, 0xAD, 0xBE, 0xEF}
		uri := "internal://binary"

		def := &configv1.ResourceDefinition{
			Uri:      &uri,
			Name:     strPtr("Binary Resource"),
			MimeType: strPtr("application/octet-stream"),
			ResourceType: &configv1.ResourceDefinition_Static{
				Static: &configv1.StaticResource{
					ContentType: &configv1.StaticResource_BinaryContent{
						BinaryContent: binaryContent,
					},
				},
			},
		}

		r := NewStaticResource(def, "test-service")

		res, err := r.Read(context.Background())

		require.NoError(t, err)
		require.NotNil(t, res)
		require.Len(t, res.Contents, 1)
		assert.Equal(t, binaryContent, res.Contents[0].Blob)
	})
}
