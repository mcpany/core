Ah, `gocritic` and `staticcheck` and `govet` are enabled.
If I disable `gocritic` and `staticcheck`? No, those are useful.
What if I use `GOGC=off` for `golangci-lint` to prevent GC from running and pausing? No, that increases memory.
What if I use `GOMEMLIMIT=2000MiB`? Yes! Go 1.19+ supports `GOMEMLIMIT` to force garbage collection before OOM!
I will add `export GOMEMLIMIT=2000MiB` before `golangci-lint` runs!
