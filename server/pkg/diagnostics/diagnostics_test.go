package diagnostics

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestRun_HTTP_Success(t *testing.T) {
	// Enable local loopback for testing
	os.Setenv("MCPANY_ALLOW_LOOPBACK_RESOURCES", "true")
	defer os.Unsetenv("MCPANY_ALLOW_LOOPBACK_RESOURCES")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	config := &configv1.UpstreamServiceConfig{
		Name: proto.String("test-service"),
		ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
			HttpService: &configv1.HttpUpstreamService{
				Address: proto.String(server.URL),
			},
		},
	}

	report := Run(context.Background(), config)

	assert.Equal(t, "test-service", report.ServiceName)
	assert.Equal(t, StatusSuccess, report.Overall)
	assert.Len(t, report.Steps, 4) // Parse, DNS, TCP, HTTP

	// Verify Steps
	assert.Equal(t, "Parse Configuration", report.Steps[0].Name)
	assert.Equal(t, StatusSuccess, report.Steps[0].Status)

	assert.Equal(t, "DNS Resolution", report.Steps[1].Name)
	assert.Equal(t, StatusSuccess, report.Steps[1].Status)

	assert.Equal(t, "TCP Connectivity", report.Steps[2].Name)
	assert.Equal(t, StatusSuccess, report.Steps[2].Status)

	assert.Equal(t, "HTTP Check", report.Steps[3].Name)
	assert.Equal(t, StatusSuccess, report.Steps[3].Status)
}

func TestRun_HTTP_InvalidURL(t *testing.T) {
	config := &configv1.UpstreamServiceConfig{
		Name: proto.String("test-service"),
		ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
			HttpService: &configv1.HttpUpstreamService{
				Address: proto.String("::invalid-url"),
			},
		},
	}

	report := Run(context.Background(), config)

	assert.Equal(t, StatusFailed, report.Overall)
	assert.Equal(t, StatusFailed, report.Steps[0].Status) // Parse failed
}

func TestRun_HTTP_Unreachable(t *testing.T) {
	// Enable local loopback for testing
	os.Setenv("MCPANY_ALLOW_LOOPBACK_RESOURCES", "true")
	defer os.Unsetenv("MCPANY_ALLOW_LOOPBACK_RESOURCES")

	listener, _ := net.Listen("tcp", "127.0.0.1:0")
	addr := listener.Addr().String()
	listener.Close()

	config := &configv1.UpstreamServiceConfig{
		Name: proto.String("test-service"),
		ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
			HttpService: &configv1.HttpUpstreamService{
				Address: proto.String("http://" + addr),
			},
		},
	}

	report := Run(context.Background(), config)

	// It might fail at TCP Connectivity
	assert.Equal(t, StatusFailed, report.Overall)

	// Parse OK
	assert.Equal(t, StatusSuccess, report.Steps[0].Status)
	// DNS OK (IP)
	assert.Equal(t, StatusSuccess, report.Steps[1].Status)
	// TCP Fail
	assert.Equal(t, StatusFailed, report.Steps[2].Status)
}
