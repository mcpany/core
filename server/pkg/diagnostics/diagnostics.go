package diagnostics

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/util"
)

type StepStatus string

const (
	StatusPending StepStatus = "pending"
	StatusRunning StepStatus = "running"
	StatusSuccess StepStatus = "success"
	StatusFailed  StepStatus = "failed"
	StatusSkipped StepStatus = "skipped"
)

type DiagnosticStep struct {
	Name       string     `json:"name"`
	Status     StepStatus `json:"status"`
	Message    string     `json:"message,omitempty"`
	Details    string     `json:"details,omitempty"`
	DurationMs int64      `json:"duration_ms"`
}

type DiagnosticReport struct {
	ServiceName string            `json:"service_name"`
	Timestamp   time.Time         `json:"timestamp"`
	Steps       []*DiagnosticStep `json:"steps"`
	Overall     StepStatus        `json:"overall"`
}

func NewReport(serviceName string) *DiagnosticReport {
	return &DiagnosticReport{
		ServiceName: serviceName,
		Timestamp:   time.Now(),
		Steps:       make([]*DiagnosticStep, 0),
		Overall:     StatusPending,
	}
}

func (r *DiagnosticReport) AddStep(name string) *DiagnosticStep {
	step := &DiagnosticStep{
		Name:   name,
		Status: StatusPending,
	}
	r.Steps = append(r.Steps, step)
	return step
}

func Run(ctx context.Context, config *configv1.UpstreamServiceConfig) *DiagnosticReport {
	report := NewReport(config.GetName())
	report.Overall = StatusRunning

	// check if it's an HTTP service
	if httpSvc := config.GetHttpService(); httpSvc != nil {
		runHTTPDiagnostics(ctx, report, httpSvc.GetAddress())
	} else if config.GetGraphqlService() != nil {
		runHTTPDiagnostics(ctx, report, config.GetGraphqlService().GetAddress())
	} else {
		// Not supported yet
		step := report.AddStep("Service Type Check")
		step.Status = StatusSkipped
		step.Message = "Diagnostics only supported for HTTP/GraphQL services currently."
		report.Overall = StatusSuccess // Not a failure of the service, just tooling limit
	}

	// Calculate overall status
	failed := false
	for _, step := range report.Steps {
		if step.Status == StatusFailed {
			failed = true
			break
		}
	}
	if failed {
		report.Overall = StatusFailed
	} else {
		report.Overall = StatusSuccess
	}

	return report
}

func runHTTPDiagnostics(ctx context.Context, report *DiagnosticReport, address string) {
	// Step 1: Parse URL
	stepParse := report.AddStep("Parse Configuration")
	start := time.Now()
	u, err := url.Parse(address)
	stepParse.DurationMs = time.Since(start).Milliseconds()

	if err != nil {
		stepParse.Status = StatusFailed
		stepParse.Message = fmt.Sprintf("Invalid URL format: %v", err)
		return
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		stepParse.Status = StatusFailed
		stepParse.Message = fmt.Sprintf("Invalid scheme: %s (must be http or https)", u.Scheme)
		return
	}
	stepParse.Status = StatusSuccess
	stepParse.Details = fmt.Sprintf("Host: %s, Port: %s, Path: %s", u.Hostname(), u.Port(), u.Path)

	// Step 2: DNS Resolution
	stepDNS := report.AddStep("DNS Resolution")
	start = time.Now()
	host := u.Hostname()

	// Handle raw IP
	if net.ParseIP(host) != nil {
		stepDNS.Status = StatusSuccess
		stepDNS.Message = "Hostname is already an IP address"
		stepDNS.Details = host
		stepDNS.DurationMs = time.Since(start).Milliseconds()
	} else {
		// Use default resolver
		ips, err := net.LookupIP(host)
		stepDNS.DurationMs = time.Since(start).Milliseconds()
		if err != nil {
			stepDNS.Status = StatusFailed
			stepDNS.Message = fmt.Sprintf("Failed to resolve host %q: %v", host, err)
			return // Stop here
		}
		var ipStrs []string
		for _, ip := range ips {
			ipStrs = append(ipStrs, ip.String())
		}
		stepDNS.Status = StatusSuccess
		stepDNS.Details = fmt.Sprintf("Resolved IPs: %s", strings.Join(ipStrs, ", "))
	}

	// Step 3: TCP Connectivity
	stepTCP := report.AddStep("TCP Connectivity")
	start = time.Now()

	// Determine port
	port := u.Port()
	if port == "" {
		if u.Scheme == "https" {
			port = "443"
		} else {
			port = "80"
		}
	}
	target := net.JoinHostPort(host, port)

	// Use SafeDialer to respect security policies
	dialer := util.NewSafeDialer()

	if envIs("MCPANY_DANGEROUS_ALLOW_LOCAL_IPS") || envIs("MCPANY_ALLOW_LOOPBACK_RESOURCES") {
		dialer.AllowLoopback = true
		dialer.AllowPrivate = true
	}
	if envIs("MCPANY_ALLOW_PRIVATE_NETWORK_RESOURCES") {
		dialer.AllowPrivate = true
	}
	dialer.Dialer = &net.Dialer{Timeout: 5 * time.Second}

	conn, err := dialer.DialContext(ctx, "tcp", target)
	stepTCP.DurationMs = time.Since(start).Milliseconds()

	if err != nil {
		stepTCP.Status = StatusFailed
		stepTCP.Message = fmt.Sprintf("Failed to connect to %s", target)
		stepTCP.Details = err.Error()
		return
	}
	conn.Close()
	stepTCP.Status = StatusSuccess
	stepTCP.Details = fmt.Sprintf("Successfully connected to %s", target)

	// Step 4: HTTP Handshake
	stepHTTP := report.AddStep("HTTP Check")
	start = time.Now()

	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			// Use the same dialer config
			DialContext:         dialer.DialContext,
			TLSHandshakeTimeout: 5 * time.Second,
		},
	}

	// Try HEAD first
	req, err := http.NewRequestWithContext(ctx, "HEAD", address, nil)
	if err != nil {
		stepHTTP.Status = StatusFailed
		stepHTTP.Message = fmt.Sprintf("Failed to create request: %v", err)
		return
	}

	// Add User-Agent
	req.Header.Set("User-Agent", "MCP-Any-Diagnostics/1.0")

	resp, err := client.Do(req)

	// If HEAD fails (e.g. 405), try GET
	if err == nil && resp.StatusCode == http.StatusMethodNotAllowed {
		resp.Body.Close()
		req, _ = http.NewRequestWithContext(ctx, "GET", address, nil)
		req.Header.Set("User-Agent", "MCP-Any-Diagnostics/1.0")
		resp, err = client.Do(req)
	}

	stepHTTP.DurationMs = time.Since(start).Milliseconds()

	if err != nil {
		stepHTTP.Status = StatusFailed
		stepHTTP.Message = "HTTP Request Failed"
		stepHTTP.Details = err.Error()
		return
	}
	defer resp.Body.Close()

	stepHTTP.Status = StatusSuccess
	stepHTTP.Details = fmt.Sprintf("Status: %s, Protocol: %s", resp.Status, resp.Proto)

	if resp.StatusCode >= 400 {
		stepHTTP.Message = fmt.Sprintf("Server returned error: %s", resp.Status)
		if resp.StatusCode >= 500 {
			stepHTTP.Status = StatusFailed
		}
	}
}

func envIs(key string) bool {
	val := os.Getenv(key)
	return val == "true" || val == "1"
}
