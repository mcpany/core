package app

import (
	"bytes"
	"context"
	"crypto/subtle"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	pb_admin "github.com/mcpany/core/proto/admin/v1"
	v1 "github.com/mcpany/core/proto/api/v1"
	"github.com/mcpany/core/server/pkg/admin"
	"github.com/mcpany/core/server/pkg/alerts"
	"github.com/mcpany/core/server/pkg/appconsts"
	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/bus"
	"github.com/mcpany/core/server/pkg/catalog"
	"github.com/mcpany/core/server/pkg/config"
	"github.com/mcpany/core/server/pkg/discovery"
	"github.com/mcpany/core/server/pkg/gc"
	"github.com/mcpany/core/server/pkg/health"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/mcpserver"
	"github.com/mcpany/core/server/pkg/metrics"
	"github.com/mcpany/core/server/pkg/middleware"
	"github.com/mcpany/core/server/pkg/pool"
	"github.com/mcpany/core/server/pkg/profile"
	"github.com/mcpany/core/server/pkg/prompt"
	"github.com/mcpany/core/server/pkg/resource"
	"github.com/mcpany/core/server/pkg/serviceregistry"
	"github.com/mcpany/core/server/pkg/skill"
	"github.com/mcpany/core/server/pkg/storage"
	"github.com/mcpany/core/server/pkg/storage/postgres"
	"github.com/mcpany/core/server/pkg/storage/sqlite"
	"github.com/mcpany/core/server/pkg/telemetry"
	"github.com/mcpany/core/server/pkg/tokenizer"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/mcpany/core/server/pkg/upstream/factory"
	"github.com/mcpany/core/server/pkg/util"
	"github.com/mcpany/core/server/pkg/validation"
	"github.com/mcpany/core/server/pkg/webhooks"
	"github.com/mcpany/core/server/pkg/worker"
	"github.com/pmezard/go-difflib/difflib"
	otelgrpc "go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	// config_v1 "github.com/mcpany/core/proto/config/v1".
	config_v1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/api/rest"
	"github.com/mcpany/core/server/pkg/topology"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/afero"
	gogrpc "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// createAuthMiddleware creates the authentication middleware.
func (a *Application) createAuthMiddleware(forcePrivateIPOnly bool, trustProxy bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Allow login endpoint without auth
			if r.URL.Path == "/api/v1/auth/login" {
				next.ServeHTTP(w, r)
				return
			}

			// Support passing Authorization via query parameter (essential for WebSockets)
			if r.Header.Get("Authorization") == "" {
				if authToken := r.URL.Query().Get("auth_token"); authToken != "" {
					// We assume Basic auth for now as that's what the UI uses for user login.
					// If the token doesn't start with "Basic " or "Bearer ", prepend "Basic ".
					if !strings.HasPrefix(authToken, "Basic ") && !strings.HasPrefix(authToken, "Bearer ") {
						r.Header.Set("Authorization", "Basic "+authToken)
					} else {
						r.Header.Set("Authorization", authToken)
					}
				}
			}

			ip := util.GetClientIP(r, trustProxy)
			ctx := util.ContextWithRemoteIP(r.Context(), ip)
			r = r.WithContext(ctx)
			apiKey := a.SettingsManager.GetAPIKey()
			requestKey := r.Header.Get("X-API-Key")
			logging.GetLogger().Info("DEBUG: AuthMiddleware details", "configured_key", apiKey, "request_key", requestKey, "path", r.URL.Path)
			authenticated := false

			// 1. Check Global API Key
			if apiKey != "" {
				requestKey := r.Header.Get("X-API-Key")
				if requestKey == "" {
					requestKey = r.URL.Query().Get("api_key")
				}
				if requestKey == "" {
					authHeader := r.Header.Get("Authorization")
					if strings.HasPrefix(authHeader, "Bearer ") {
						requestKey = strings.TrimPrefix(authHeader, "Bearer ")
					}
				}

				if subtle.ConstantTimeCompare([]byte(requestKey), []byte(apiKey)) == 1 {
					authenticated = true
					// Inject API Key into context if needed
					ctx = auth.ContextWithAPIKey(ctx, requestKey)
					// Global API Key grants Admin privileges (Root Access)
					ctx = auth.ContextWithRoles(ctx, []string{"admin"})
					// Also inject a placeholder user ID so that handlers expecting a user context don't fail
					ctx = auth.ContextWithUser(ctx, "system-admin")
				}
			}

			// 2. Check User Authentication (Basic Auth)
			if !authenticated {
				username, _, ok := r.BasicAuth()
				if ok && a.AuthManager != nil {
					if user, found := a.AuthManager.GetUser(username); found {
						if err := auth.ValidateAuthentication(ctx, user.GetAuthentication(), r); err == nil {
							authenticated = true
							ctx = auth.ContextWithUser(ctx, username)
							if len(user.GetRoles()) > 0 {
								ctx = auth.ContextWithRoles(ctx, user.GetRoles())
							}
						}
					}
				}
			}

			if authenticated {
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			if !forcePrivateIPOnly && apiKey != "" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Sentinel Security: If no API key is configured (and no user auth succeeded), enforce localhost-only access.
			// This prevents accidental exposure of the server to the public internet (RCE risk).
			if apiKey == "" {
				host, _, err := net.SplitHostPort(r.RemoteAddr)
				if err != nil {
					// Fallback if RemoteAddr is weird, assume host is the string itself
					host = r.RemoteAddr
				}

				// Check if the request is from a loopback address
				ipAddr := net.ParseIP(host)
				if !util.IsPrivateIP(ipAddr) {
					logging.GetLogger().Warn("Blocked public internet request because no API Key is configured", "remote_addr", r.RemoteAddr)
					http.Error(w, "Forbidden: Public access requires an API Key to be configured", http.StatusForbidden)
					return
				}

				// Grant Admin privileges (Root Access) for local development/testing convenience
				// when running in insecure mode (private network, no API key).
				ctx = auth.ContextWithRoles(ctx, []string{"admin"})
				ctx = auth.ContextWithUser(ctx, "system-admin")
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// HTTPRequestContextMiddleware injects the HTTP request into the context.
//
// Summary: Middleware to add HTTP request to context.
//
// Parameters:
//   - next (http.Handler): The next handler.
//
// Returns:
//   - (http.Handler): The wrapped handler.
func (a *Application) HTTPRequestContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), middleware.HTTPRequestContextKey, r)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetAuditMiddleware returns the current audit middleware.
//
// Summary: Returns the active audit middleware.
//
// Returns:
//   - *middleware.AuditMiddleware: The current audit middleware instance.
func (a *Application) GetAuditMiddleware() *middleware.AuditMiddleware {
	a.configMu.Lock()
	defer a.configMu.Unlock()
	if a.standardMiddlewares != nil {
		return a.standardMiddlewares.Audit
	}
	return nil
}
