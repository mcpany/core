// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

//go:build e2e

package e2e

import (
	"os/exec"
	"strings"
	"testing"

	"gopkg.in/yaml.v2"
)

type Config struct {
	UpstreamServices []struct {
		Name        string `yaml:"name"`
		HTTPService struct {
			Address string `yaml:"address"`
			Calls   []struct {
				OperationID  string `yaml:"operationId"`
				Description  string `yaml:"description"`
				Method       string `yaml:"method"`
				EndpointPath string `yaml:"endpointPath"`
			} `yaml:"calls"`
		} `yaml:"httpService"`
	} `yaml:"upstreamServices"`
}

func TestCLI(t *testing.T) {
	cmd := exec.Command("go", "run", "../../cmd/server/main.go", "config", "generate")

	input := "http\nmy-service\nhttp://example.com\nget-user\nGet a user\nHTTP_METHOD_GET\n/users/{id}\n"
	cmd.Stdin = strings.NewReader(input)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("unexpected error: %v\noutput: %s", err, string(output))
	}

	outputParts := strings.Split(string(output), "Generated configuration:")
	if len(outputParts) != 2 {
		t.Fatalf("unexpected output format: %s", string(output))
	}

	var cfg Config
	err = yaml.Unmarshal([]byte(outputParts[1]), &cfg)
	if err != nil {
		t.Fatalf("failed to unmarshal YAML: %v", err)
	}

	if len(cfg.UpstreamServices) != 1 {
		t.Fatalf("expected 1 upstream service, but got %d", len(cfg.UpstreamServices))
	}

	service := cfg.UpstreamServices[0]
	if service.Name != "my-service" {
		t.Errorf("unexpected service name: got %s, want my-service", service.Name)
	}

	if service.HTTPService.Address != "http://example.com" {
		t.Errorf("unexpected service address: got %s, want http://example.com", service.HTTPService.Address)
	}

	if len(service.HTTPService.Calls) != 1 {
		t.Fatalf("expected 1 call, but got %d", len(service.HTTPService.Calls))
	}

	call := service.HTTPService.Calls[0]
	if call.OperationID != "get-user" {
		t.Errorf("unexpected operation ID: got %s, want get-user", call.OperationID)
	}

	if call.Description != "Get a user" {
		t.Errorf("unexpected description: got %s, want Get a user", call.Description)
	}

	if call.Method != "HTTP_METHOD_GET" {
		t.Errorf("unexpected method: got %s, want HTTP_METHOD_GET", call.Method)
	}

	if call.EndpointPath != "/users/{id}" {
		t.Errorf("unexpected endpoint path: got %s, want /users/{id}", call.EndpointPath)
	}
}
