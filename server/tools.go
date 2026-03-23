// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

//go:build tools
// +build tools

// Package main provides toolchain dependencies.
package main

import (
	_ "github.com/go-logr/zapr"
	_ "google.golang.org/genproto/googleapis/api/annotations"
	_ "google.golang.org/grpc/cmd/protoc-gen-go-grpc"
	_ "google.golang.org/protobuf/cmd/protoc-gen-go"
	_ "k8s.io/api/core/v1"
	_ "k8s.io/apimachinery/pkg/runtime"
	_ "sigs.k8s.io/controller-runtime/pkg/client"
)
