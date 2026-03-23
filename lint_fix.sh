# Remove the unused variable in public_api tests if we missed any
cd server
$(go env GOPATH)/bin/golangci-lint run ./... > lint.log 2>&1
cat lint.log | grep -v 'could not import' | grep -v 'unsupported version: 2'
