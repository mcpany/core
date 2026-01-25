#!/bin/bash
PACKAGES=$(go list ./cmd/... ./pkg/... ./tests/... ./examples/upstream_service_demo/... ./docs/... | \
    grep -v /tests/public_api | \
    grep -v /pkg/command | \
    grep -v /build | \
    grep -v /tests/e2e_sequential | \
    sort)

SHARD_INDEX=3
SHARD_TOTAL=4
COUNT=0
for PKG in $PACKAGES; do
    if [ $(( (COUNT % SHARD_TOTAL) + 1 )) -eq "$SHARD_INDEX" ]; then
        echo "$PKG"
    fi
    COUNT=$((COUNT + 1))
done
