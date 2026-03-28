//go:build tools
package util
import (
    _ "github.com/go-logr/zapr"
    _ "google.golang.org/grpc"
    _ "google.golang.org/genproto/googleapis/api/annotations"
    _ "k8s.io/api/core/v1"
    _ "k8s.io/apimachinery/pkg/runtime"
    _ "sigs.k8s.io/controller-runtime"
    _ "go.opentelemetry.io/contrib/detectors/gcp"
    _ "github.com/envoyproxy/go-control-plane/envoy/service/status/v3"
)
