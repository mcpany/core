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

// configHealthCheck checks the status of the configuration.
func (a *Application) configHealthCheck(_ context.Context) health.CheckResult {
	a.configMu.Lock()
	defer a.configMu.Unlock()

	if a.lastReloadErr != nil {
		return health.CheckResult{
			Status:  "degraded",
			Message: a.lastReloadErr.Error(),
			Latency: time.Since(a.lastReloadTime).String(),
			Diff:    a.configDiff,
		}
	}

	status := "ok"
	if a.lastReloadTime.IsZero() {
		status = "unknown"
	}

	return health.CheckResult{
		Status:  status,
		Latency: time.Since(a.lastReloadTime).String(),
	}
}

func (a *Application) filesystemHealthCheck(_ context.Context) health.CheckResult {
	if a.ServiceRegistry == nil {
		return health.CheckResult{Status: "ok"}
	}

	services, err := a.ServiceRegistry.GetAllServices()
	if err != nil {
		return health.CheckResult{
			Status:  "degraded",
			Message: fmt.Sprintf("failed to list services: %v", err),
		}
	}

	var issues []string
	start := time.Now()

	for _, svc := range services {
		fsSvc := svc.GetFilesystemService()
		if fsSvc == nil {
			continue
		}

		for virtualPath, localPath := range fsSvc.GetRootPaths() {
			if info, err := os.Stat(localPath); err != nil {
				issues = append(issues, fmt.Sprintf("service %q: root path %q (%s) is inaccessible: %v", svc.GetName(), virtualPath, localPath, err))
			} else if !info.IsDir() {
				issues = append(issues, fmt.Sprintf("service %q: root path %q (%s) is not a directory", svc.GetName(), virtualPath, localPath))
			}
		}
	}

	status := "ok"
	var message string
	if len(issues) > 0 {
		status = "degraded"
		message = strings.Join(issues, "; ")
	}

	return health.CheckResult{
		Status:  status,
		Message: message,
		Latency: time.Since(start).String(),
	}
}
