export GOGC=off
export GOMEMLIMIT=1500MiB
cd server
../build/env/bin/golangci-lint run --timeout=5m -j 2 --issues-exit-code 1
