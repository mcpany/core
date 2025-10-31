/*
 * Copyright 2025 Author(s) of MCP Any
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package util

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"

	configv1 "github.com/mcpany/core/proto/config/v1"
)

// NewHttpClientWithTLS creates a new *http.Client configured with the specified
// TLS settings. It supports setting a custom CA certificate, a client
// certificate and key, the server name for SNI, and skipping verification. If
// the provided tlsConfig is nil, it returns http.DefaultClient.
//
// tlsConfig contains the TLS settings to apply to the HTTP client's transport.
// It returns a configured *http.Client or an error if the TLS configuration
// is invalid or files cannot be read.
func NewHttpClientWithTLS(tlsConfig *configv1.TLSConfig) (*http.Client, error) {
	if tlsConfig == nil {
		return http.DefaultClient, nil
	}

	config := &tls.Config{
		ServerName:         tlsConfig.GetServerName(),
		InsecureSkipVerify: tlsConfig.GetInsecureSkipVerify(),
	}

	if tlsConfig.GetCaCertPath() != "" {
		caCert, err := os.ReadFile(tlsConfig.GetCaCertPath())
		if err != nil {
			return nil, fmt.Errorf("failed to read CA certificate: %w", err)
		}
		caCertPool := x509.NewCertPool()
		if ok := caCertPool.AppendCertsFromPEM(caCert); !ok {
			return nil, fmt.Errorf("failed to append CA certs from PEM")
		}
		config.RootCAs = caCertPool
	}

	if tlsConfig.GetClientCertPath() != "" && tlsConfig.GetClientKeyPath() != "" {
		clientCert, err := tls.LoadX509KeyPair(tlsConfig.GetClientCertPath(), tlsConfig.GetClientKeyPath())
		if err != nil {
			return nil, fmt.Errorf("failed to load client key pair: %w", err)
		}
		config.Certificates = []tls.Certificate{clientCert}
	}

	transport := &http.Transport{
		TLSClientConfig: config,
	}

	return &http.Client{
		Transport: transport,
	}, nil
}
