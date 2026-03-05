// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

//go:build tools
// +build tools

package main

import (
	_ "google.golang.org/grpc/cmd/protoc-gen-go-grpc"
	_ "google.golang.org/protobuf/cmd/protoc-gen-go"
)
