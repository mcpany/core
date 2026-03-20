#!/bin/bash
cd server
rm -rf /tmp/go1.26
mkdir -p /tmp/go1.26
wget -qO- https://go.dev/dl/go1.26.1.linux-amd64.tar.gz | tar -xzC /tmp/go1.26 --strip-components=1
export PATH=/tmp/go1.26/bin:$PATH
export GOTOOLCHAIN=local
/home/jules/go/bin/golangci-lint run ./cmd/server/...
