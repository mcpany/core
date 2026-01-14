// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package diagnostics provides functionality for gathering server diagnostics.
package diagnostics

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/mcpany/core/server/pkg/appconsts"
	"github.com/mcpany/core/server/pkg/config"
	"github.com/mcpany/core/server/pkg/serviceregistry"
	"github.com/shirou/gopsutil/v3/mem"
)

// Service collects diagnostic information about the server.
type Service struct {
	startTime       time.Time
	serviceRegistry serviceregistry.Interface
}

// NewService creates a new DiagnosticsService.
func NewService(serviceRegistry serviceregistry.Interface) *Service {
	return &Service{
		startTime:       time.Now(),
		serviceRegistry: serviceRegistry,
	}
}

// SystemInfo contains system-level information.
type SystemInfo struct {
	Version      string `json:"version"`
	GoVersion    string `json:"go_version"`
	OS           string `json:"os"`
	Arch         string `json:"arch"`
	Uptime       string `json:"uptime"`
	UptimeSeconds float64 `json:"uptime_seconds"`
	Memory       string `json:"memory_usage"` // Human readable
	NumGoroutine int    `json:"num_goroutine"`
}

// ConfigInfo contains information about the loaded configuration.
type ConfigInfo struct {
	ConfigPaths []string `json:"config_paths"`
}

// ServiceStatus contains status of an upstream service.
type ServiceStatus struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Status   string `json:"status"` // "Registered", "Error", "Unknown"
	Error    string `json:"error,omitempty"`
}

// DiagnosticReport is the full diagnostic report.
type DiagnosticReport struct {
	System   SystemInfo      `json:"system"`
	Config   ConfigInfo      `json:"config"`
	Services []ServiceStatus `json:"services"`
	Timestamp string         `json:"timestamp"`
}

// GenerateReport generates a diagnostic report.
func (s *Service) GenerateReport(_ context.Context) (*DiagnosticReport, error) {
	report := &DiagnosticReport{
		Timestamp: time.Now().Format(time.RFC3339),
	}

	// System Info
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	v, _ := mem.VirtualMemory()
	memUsage := "Unknown"
	if v != nil {
		memUsage = fmt.Sprintf("%v MB", v.Used/1024/1024)
	}

	report.System = SystemInfo{
		Version:       appconsts.Version,
		GoVersion:     runtime.Version(),
		OS:            runtime.GOOS,
		Arch:          runtime.GOARCH,
		Uptime:        time.Since(s.startTime).String(),
		UptimeSeconds: time.Since(s.startTime).Seconds(),
		Memory:        memUsage,
		NumGoroutine:  runtime.NumGoroutine(),
	}

	// Config Info
	report.Config = ConfigInfo{
		ConfigPaths: config.GlobalSettings().ConfigPaths(),
	}

	// Services Info
	report.Services = []ServiceStatus{} // Ensure empty slice instead of nil for JSON
	if s.serviceRegistry != nil {
		services, _ := s.serviceRegistry.GetAllServices()
		for _, svc := range services {
			id := svc.GetName() // Assuming ID is Name for now or we should get sanitized ID
			status := "Registered"

			// Check if there is a registration error
			errMsg, hasError := s.serviceRegistry.GetServiceError(id)
			if hasError {
				status = "Error"
			}

			// Determine type
			// Simple mapping based on the enum or type
			typeStr := fmt.Sprintf("%T", svc.WhichServiceConfig())

			report.Services = append(report.Services, ServiceStatus{
				ID:     id,
				Name:   svc.GetName(),
				Type:   typeStr,
				Status: status,
				Error:  errMsg,
			})
		}
	}

	return report, nil
}
