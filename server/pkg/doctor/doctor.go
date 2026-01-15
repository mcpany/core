// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package doctor

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/mcpany/core/server/pkg/config"
	"github.com/spf13/afero"
)

// Severity represents the severity of a check result.
type Severity int

const (
	// Info indicates a purely informational message.
	Info Severity = iota
	// Warning indicates a potential issue that doesn't prevent startup but might degrade functionality.
	Warning
	// Error indicates a critical issue that will likely prevent startup or major functionality.
	Error
)

// String returns the string representation of the severity.
func (s Severity) String() string {
	switch s {
	case Info:
		return "INFO"
	case Warning:
		return "WARN"
	case Error:
		return "FAIL"
	default:
		return "UNKNOWN"
	}
}

// Result represents the result of a single check.
type Result struct {
	Name        string
	Severity    Severity
	Message     string
	Remediation string
	Error       error
}

// Check is the interface that all diagnostics must implement.
type Check interface {
	// Name returns the unique name of the check.
	Name() string
	// Run executes the check and returns a Result.
	Run(ctx context.Context) Result
}

// Doctor manages and runs a suite of diagnostic checks.
type Doctor struct {
	checks []Check
}

// NewDoctor creates a new Doctor instance with default checks.
func NewDoctor(cfg *config.Settings, fs afero.Fs) *Doctor {
	d := &Doctor{
		checks: []Check{
			&ConfigCheck{fs: fs, paths: cfg.ConfigPaths()},
			&PortCheck{host: cfg.MCPListenAddress()},
			// We can add more checks here
		},
	}
	// Add GRPC port check if configured
	if cfg.GRPCPort() != "" {
		d.checks = append(d.checks, &PortCheck{host: cfg.GRPCPort(), nameSuffix: "(gRPC)"})
	}

	return d
}

// Diagnose runs all registered checks and returns the results.
func (d *Doctor) Diagnose(ctx context.Context) []Result {
	var results []Result
	for _, check := range d.checks {
		results = append(results, check.Run(ctx))
	}
	return results
}

// --- Implementations ---

// ConfigCheck validates the configuration files.
type ConfigCheck struct {
	fs    afero.Fs
	paths []string
}

func NewConfigCheck(fs afero.Fs, paths []string) *ConfigCheck {
	return &ConfigCheck{fs: fs, paths: paths}
}

func (c *ConfigCheck) Name() string {
	return "Configuration Check"
}

func (c *ConfigCheck) Run(ctx context.Context) Result {
	if len(c.paths) == 0 {
		return Result{
			Name:     c.Name(),
			Severity: Warning,
			Message:  "No configuration files provided.",
			Remediation: "Pass config files via --config-path or env vars if you intend to load services.",
		}
	}

	store := config.NewFileStore(c.fs, c.paths)
	configs, err := config.LoadResolvedConfig(ctx, store)
	if err != nil {
		return Result{
			Name:        c.Name(),
			Severity:    Error,
			Message:     fmt.Sprintf("Failed to load configuration: %v", err),
			Remediation: "Check if the file exists and is valid YAML/JSON.",
			Error:       err,
		}
	}

	validationErrors := config.Validate(ctx, configs, config.Server)
	if len(validationErrors) > 0 {
		var msgs []string
		for _, e := range validationErrors {
			msgs = append(msgs, e.Error())
		}
		return Result{
			Name:        c.Name(),
			Severity:    Error,
			Message:     fmt.Sprintf("Configuration validation failed:\n  - %s", strings.Join(msgs, "\n  - ")),
			Remediation: "Fix the validation errors in your config file.",
			Error:       fmt.Errorf("validation failed"),
		}
	}

	return Result{
		Name:     c.Name(),
		Severity: Info,
		Message:  fmt.Sprintf("Successfully loaded and validated %d configuration file(s).", len(c.paths)),
	}
}

// PortCheck verifies if the port is available.
type PortCheck struct {
	host       string
	nameSuffix string
}

func NewPortCheck(host string) *PortCheck {
	return &PortCheck{host: host}
}

func (c *PortCheck) Name() string {
	name := "Port Availability"
	if c.nameSuffix != "" {
		name += " " + c.nameSuffix
	}
	return name
}

func (c *PortCheck) Run(ctx context.Context) Result {
	host := c.host
	if host == "" {
		return Result{
			Name:     c.Name(),
			Severity: Info,
			Message:  "No port configured to check.",
		}
	}

	// If no port specified in string (e.g. just ":8080" or "8080" or "localhost:8080")
	// If it doesn't have colon, assume it's just a port if it's a number?
	// The config usually expects "host:port" or ":port".
	if !strings.Contains(host, ":") {
		// If it looks like a port number, prepend :
		host = ":" + host
	}

	// Try to bind to the port
	ln, err := net.Listen("tcp", host)
	if err != nil {
		msg := fmt.Sprintf("Port %s is not available: %v", host, err)
		severity := Error
		remediation := fmt.Sprintf("Identify the process using this port (e.g. `lsof -i %s`) and kill it, or choose a different port.", host)

		// Special handling for permission denied (port < 1024)
		if strings.Contains(err.Error(), "permission denied") {
			remediation = "Run with sudo or choose a port > 1024."
		}

		return Result{
			Name:        c.Name(),
			Severity:    severity,
			Message:     msg,
			Remediation: remediation,
			Error:       err,
		}
	}
	_ = ln.Close()

	return Result{
		Name:     c.Name(),
		Severity: Info,
		Message:  fmt.Sprintf("Port %s is available.", host),
	}
}

// EnvCheck checks if critical environment variables are set.
// This assumes we can infer required env vars from config placeholders or standard ones.
// For now, we'll check for GITHUB_TOKEN if we see github services?
// Or maybe just generic check.
// Let's keep it simple for now and rely on config validation which should handle missing env vars if they are required.
