Change `toolchain go1.26.1` to `toolchain go1.24.0` in `server/go.mod` because golangci-lint v1.64.5 doesn't support go1.26.1, which triggers a panic/failure during goanalysis_metalinter.
Wait, let's also look at `scripts/lint.sh`.
