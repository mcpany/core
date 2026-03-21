#!/bin/bash
export PATH=$(go env GOPATH)/bin:$PATH
golangci-lint run ./server/pkg/app/api.go
