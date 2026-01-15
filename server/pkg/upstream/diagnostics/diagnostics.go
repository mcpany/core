// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package diagnostics provides tools for diagnosing connection issues with upstream services.
package diagnostics

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os/exec"
	"strings"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
)

const (
	statusSuccess = "success"
	statusFailure = "failure"
	statusRunning = "running"
)

// DiagnosticStep represents a single step in the diagnostic process.
type DiagnosticStep struct {
	Name    string `json:"name"`
	Status  string `json:"status"` // "pending", "running", "success", "failure", "skipped"
	Message string `json:"message,omitempty"`
	Latency string `json:"latency,omitempty"`
}

// DiagnosticReport represents the full diagnostic report.
type DiagnosticReport struct {
	ServiceName string           `json:"service_name"`
	Steps       []DiagnosticStep `json:"steps"`
}

// DiagnoseService runs diagnostics for the given service configuration.
func DiagnoseService(ctx context.Context, config *configv1.UpstreamServiceConfig) DiagnosticReport {
	report := DiagnosticReport{
		ServiceName: config.GetName(),
		Steps:       []DiagnosticStep{},
	}

	// 1. Basic Configuration Check
	step := DiagnosticStep{Name: "Configuration Check", Status: statusRunning}
	if config.GetName() == "" {
		step.Status = statusFailure
		step.Message = "Service name is empty"
	} else {
		step.Status = statusSuccess
		step.Message = "Configuration appears valid"
	}
	report.Steps = append(report.Steps, step)

	// Dispatch based on service type
	if httpSvc := config.GetHttpService(); httpSvc != nil {
		report.Steps = append(report.Steps, diagnoseHTTP(ctx, httpSvc)...)
	} else if grpcSvc := config.GetGrpcService(); grpcSvc != nil {
		report.Steps = append(report.Steps, diagnoseGRPC(ctx, grpcSvc)...)
	} else if mcpSvc := config.GetMcpService(); mcpSvc != nil {
		report.Steps = append(report.Steps, diagnoseMCP(ctx, mcpSvc)...)
	} else if cmdSvc := config.GetCommandLineService(); cmdSvc != nil {
		report.Steps = append(report.Steps, diagnoseCommand(cmdSvc)...)
	}

	return report
}

func diagnoseHTTP(ctx context.Context, svc *configv1.HttpUpstreamService) []DiagnosticStep {
	steps := []DiagnosticStep{}
	url := svc.GetAddress()

	// 1. DNS Resolution / Host Reachability
	steps = append(steps, checkReachability(ctx, url))

	// 2. HTTP Handshake
	step := DiagnosticStep{Name: "HTTP Handshake", Status: statusRunning}
	start := time.Now()
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		step.Status = statusFailure
		step.Message = fmt.Sprintf("Failed to create request: %v", err)
	} else {
		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			step.Status = statusFailure
			step.Message = fmt.Sprintf("Request failed: %v", err)
		} else {
			defer func() {
				_ = resp.Body.Close()
			}()
			step.Status = statusSuccess
			step.Message = fmt.Sprintf("Received status %s", resp.Status)
		}
	}
	step.Latency = time.Since(start).String()
	steps = append(steps, step)

	return steps
}

func diagnoseGRPC(ctx context.Context, svc *configv1.GrpcUpstreamService) []DiagnosticStep {
	steps := []DiagnosticStep{}
	addr := svc.GetAddress()

	// 1. TCP Connect
	steps = append(steps, checkTCPConnect(ctx, addr))

	// Note: Full gRPC handshake requires a gRPC client, skipping for now to avoid large deps
	// unless we want to use `grpc.Dial` with insecure just to check connection.

	return steps
}

func diagnoseMCP(ctx context.Context, svc *configv1.McpUpstreamService) []DiagnosticStep {
	steps := []DiagnosticStep{}

	if sse := svc.GetHttpConnection(); sse != nil {
		// HTTP/SSE Check (McpStreamableHttpConnection uses GetHttpAddress)
		steps = append(steps, checkReachability(ctx, sse.GetHttpAddress()))
		steps = append(steps, checkTCPConnect(ctx, sse.GetHttpAddress()))
	} else if stdio := svc.GetStdioConnection(); stdio != nil {
		// Command Check
		steps = append(steps, checkCommand(stdio.GetCommand()))
	}

	return steps
}

func diagnoseCommand(svc *configv1.CommandLineUpstreamService) []DiagnosticStep {
	return []DiagnosticStep{
		checkCommand(svc.GetCommand()),
	}
}

// Helpers

func checkReachability(ctx context.Context, target string) DiagnosticStep {
	step := DiagnosticStep{Name: "Reachability Check", Status: statusRunning}
	start := time.Now()

	// Extract host
	host := target
	if strings.Contains(target, "://") {
		parts := strings.Split(target, "://")
		if len(parts) > 1 {
			host = parts[1]
		}
	}
	if strings.Contains(host, "/") {
		host = strings.Split(host, "/")[0]
	}
	if strings.Contains(host, ":") {
		host, _, _ = net.SplitHostPort(host)
	}

	ips, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil {
		step.Status = statusFailure
		step.Message = fmt.Sprintf("DNS Resolution failed: %v", err)
	} else {
		step.Status = statusSuccess
		var ipStrs []string
		for _, ip := range ips {
			ipStrs = append(ipStrs, ip.String())
		}
		step.Message = fmt.Sprintf("Resolved to: %s", strings.Join(ipStrs, ", "))
	}
	step.Latency = time.Since(start).String()
	return step
}

func checkTCPConnect(ctx context.Context, target string) DiagnosticStep {
	step := DiagnosticStep{Name: "TCP Connection", Status: statusRunning}
	start := time.Now()

	host := target
	// Clean up scheme
	if strings.Contains(target, "://") {
		parts := strings.Split(target, "://")
		host = parts[1]
	}
	// Needs port
	if !strings.Contains(host, ":") {
		// Default ports if missing (heuristic)
		if strings.HasPrefix(target, "https") {
			host += ":443"
		} else {
			host += ":80"
		}
	}

	d := net.Dialer{Timeout: 5 * time.Second}
	conn, err := d.DialContext(ctx, "tcp", host)
	if err != nil {
		step.Status = statusFailure
		step.Message = fmt.Sprintf("Connection refused: %v", err)
	} else {
		_ = conn.Close()
		step.Status = statusSuccess
		step.Message = "Connection established"
	}
	step.Latency = time.Since(start).String()
	return step
}

func checkCommand(command string) DiagnosticStep {
	step := DiagnosticStep{Name: "Command Check", Status: statusRunning}
	start := time.Now()

	path, err := exec.LookPath(command)
	if err != nil {
		step.Status = statusFailure
		step.Message = fmt.Sprintf("Command not found in PATH: %v", err)
	} else {
		step.Status = statusSuccess
		step.Message = fmt.Sprintf("Found at: %s", path)
	}
	step.Latency = time.Since(start).String()
	return step
}
