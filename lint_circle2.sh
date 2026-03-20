#!/bin/bash
cd server
export PATH=$PWD/../build/env/bin:$PATH
golangci-lint run --timeout 20m ./cmd/server/...
