#!/bin/bash
export PATH=/tmp/go1.26/bin:$PATH
export GOTOOLCHAIN=auto
export GOWORK=off
cd server
git checkout origin/main -- ../go.work go.mod .golangci.yml
/home/jules/go/bin/golangci-lint run --timeout=10m ./cmd/server/...
