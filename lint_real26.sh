#!/bin/bash
export PATH=/tmp/go1.26/bin:$PATH
export GOTOOLCHAIN=auto
cd server
/home/jules/go/bin/golangci-lint run --disable-all -E revive -E gofumpt ./cmd/server/...
