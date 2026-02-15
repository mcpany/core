package resource

import (
	"context"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestStaticResource_Read_LimitOverflow(t *testing.T) {
	// Enable loopback for testing
	t.Setenv("MCPANY_ALLOW_LOOPBACK_RESOURCES", "true")

	// Setup test server returning some content
	content := "hello world"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte(content))
	}))
	defer server.Close()

	// Define a resource with MaxInt64 size.
	// This should logically allow reading everything.
	// But due to overflow in limit+1 logic, it might fail.
	limit := int64(math.MaxInt64)
	def := configv1.ResourceDefinition_builder{
		Uri:      proto.String(server.URL),
		Name:     proto.String("Max Size Resource"),
		MimeType: proto.String("text/plain"),
		Size:     &limit,
	}.Build()

	serviceID := "test-service"
	r := NewStaticResource(def, serviceID)

	// Execute Read
	res, err := r.Read(context.Background())
	require.NoError(t, err)
	require.Len(t, res.Contents, 1)

	// Verify content
	// If the bug exists, this will likely be empty or fail.
	assert.Equal(t, []byte(content), res.Contents[0].Blob, "Content should be read correctly even with MaxInt64 limit")
}
