// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"net/url"
)

func main() {
	baseURL, _ := url.Parse("http://localhost:8080?base=param")
	endpointPath := "/test?q=hello%26world"
	endpointURL, _ := url.Parse(endpointPath)

	fmt.Printf("Endpoint RawQuery: %s\n", endpointURL.RawQuery)
	fmt.Printf("Endpoint Query: %v\n", endpointURL.Query())

	resolvedURL := baseURL.JoinPath(endpointURL.Path)
	fmt.Printf("ResolvedURL after JoinPath: %s\n", resolvedURL.String())

	query := resolvedURL.Query()
	endpointQuery := endpointURL.Query()
	for k, v := range endpointQuery {
		query[k] = v
	}
	resolvedURL.RawQuery = query.Encode()
	fmt.Printf("Final URL: %s\n", resolvedURL.String())
}
