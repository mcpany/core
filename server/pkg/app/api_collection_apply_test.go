package app

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/storage/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestHandleCollectionApply(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		collectionName string
		setupStore     func(store *memory.Store)
		setupEnv       func(t *testing.T)
		expectedStatus int
		verify         func(t *testing.T, store *memory.Store)
	}{
		{
			name:           "Method Not Allowed",
			method:         http.MethodGet,
			collectionName: "test-collection",
			setupStore:     func(store *memory.Store) {},
			setupEnv:       func(t *testing.T) {},
			expectedStatus: http.StatusMethodNotAllowed,
			verify:         func(t *testing.T, store *memory.Store) {},
		},
		{
			name:           "Not Found",
			method:         http.MethodPost,
			collectionName: "nonexistent",
			setupStore:     func(store *memory.Store) {},
			setupEnv:       func(t *testing.T) {},
			expectedStatus: http.StatusNotFound,
			verify:         func(t *testing.T, store *memory.Store) {},
		},
		{
			name:           "Happy Path - Safe Service",
			method:         http.MethodPost,
			collectionName: "safe-collection",
			setupStore: func(store *memory.Store) {
				httpUpstream := configv1.HttpUpstreamService_builder{
					Address: proto.String("http://example.com"),
				}.Build()

				svc := configv1.UpstreamServiceConfig_builder{
					Name: proto.String("safe-service"),
				}.Build()
				svc.SetHttpService(httpUpstream)

				col := configv1.Collection_builder{
					Name:     proto.String("safe-collection"),
					Services: []*configv1.UpstreamServiceConfig{svc},
				}.Build()

				err := store.SaveServiceCollection(context.Background(), col)
				require.NoError(t, err)
			},
			setupEnv:       func(t *testing.T) {},
			expectedStatus: http.StatusOK,
			verify: func(t *testing.T, store *memory.Store) {
				svc, err := store.GetService(context.Background(), "safe-service")
				assert.NoError(t, err)
				assert.NotNil(t, svc)
			},
		},
		{
			name:           "Unsafe Service - Allowed",
			method:         http.MethodPost,
			collectionName: "unsafe-collection",
			setupStore: func(store *memory.Store) {
				stdioConn := configv1.McpStdioConnection_builder{
					Command: proto.String("echo"),
				}.Build()

				mcpUpstream := configv1.McpUpstreamService_builder{}.Build()
				mcpUpstream.SetStdioConnection(stdioConn)

				svc := configv1.UpstreamServiceConfig_builder{
					Name: proto.String("unsafe-service"),
				}.Build()
				svc.SetMcpService(mcpUpstream)

				col := configv1.Collection_builder{
					Name:     proto.String("unsafe-collection"),
					Services: []*configv1.UpstreamServiceConfig{svc},
				}.Build()

				err := store.SaveServiceCollection(context.Background(), col)
				require.NoError(t, err)
			},
			setupEnv: func(t *testing.T) {
				t.Setenv("MCPANY_ALLOW_UNSAFE_CONFIG", "true")
			},
			expectedStatus: http.StatusOK,
			verify: func(t *testing.T, store *memory.Store) {
				svc, err := store.GetService(context.Background(), "unsafe-service")
				assert.NoError(t, err)
				assert.NotNil(t, svc)
			},
		},
		{
			name:           "Unsafe Service - Not Allowed",
			method:         http.MethodPost,
			collectionName: "unsafe-blocked-collection",
			setupStore: func(store *memory.Store) {
				stdioConn := configv1.McpStdioConnection_builder{
					Command: proto.String("echo"),
				}.Build()

				mcpUpstream := configv1.McpUpstreamService_builder{}.Build()
				mcpUpstream.SetStdioConnection(stdioConn)

				svc := configv1.UpstreamServiceConfig_builder{
					Name: proto.String("unsafe-blocked-service"),
				}.Build()
				svc.SetMcpService(mcpUpstream)

				col := configv1.Collection_builder{
					Name:     proto.String("unsafe-blocked-collection"),
					Services: []*configv1.UpstreamServiceConfig{svc},
				}.Build()

				err := store.SaveServiceCollection(context.Background(), col)
				require.NoError(t, err)
			},
			setupEnv: func(t *testing.T) {
				t.Setenv("MCPANY_ALLOW_UNSAFE_CONFIG", "false")
			},
			expectedStatus: http.StatusOK,
			verify: func(t *testing.T, store *memory.Store) {
				svc, err := store.GetService(context.Background(), "unsafe-blocked-service")
				assert.NoError(t, err) // Expect nil here since it's GetService for non-existent service
				assert.Nil(t, svc)
			},
		},
		{
			name:           "Invalid Service in Collection",
			method:         http.MethodPost,
			collectionName: "invalid-collection",
			setupStore: func(store *memory.Store) {
				svc := configv1.UpstreamServiceConfig_builder{
					Name: proto.String("invalid-service"),
					// Missing HttpService or McpService, should fail validation
				}.Build()

				col := configv1.Collection_builder{
					Name:     proto.String("invalid-collection"),
					Services: []*configv1.UpstreamServiceConfig{svc},
				}.Build()

				err := store.SaveServiceCollection(context.Background(), col)
				require.NoError(t, err)
			},
			setupEnv:       func(t *testing.T) {},
			expectedStatus: http.StatusOK,
			verify: func(t *testing.T, store *memory.Store) {
				svc, err := store.GetService(context.Background(), "invalid-service")
				assert.NoError(t, err) // Expect nil since memory store returns nil error for not found
				assert.Nil(t, svc)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			store := memory.NewStore()
			if tc.setupStore != nil {
				tc.setupStore(store)
			}
			if tc.setupEnv != nil {
				tc.setupEnv(t)
			}

			app := &Application{}

			req := httptest.NewRequest(tc.method, "/api/v1/collections/apply", nil)
			w := httptest.NewRecorder()

			app.handleCollectionApply(w, req, tc.collectionName, store)

			assert.Equal(t, tc.expectedStatus, w.Code)
			if tc.verify != nil {
				tc.verify(t, store)
			}
		})
	}
}
