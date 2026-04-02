cd server
go mod edit -replace github.com/mcpany/core/proto=../proto
go mod tidy
go test -v -count=1 -tags=e2e ./docs/features/caching/... || echo "test failed"
