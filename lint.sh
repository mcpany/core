#!/bin/bash
set -e

# Run standard go fmt and go vet
cd server
go fmt ./pkg/tool/...
go vet ./pkg/tool/...
