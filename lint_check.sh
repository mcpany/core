cd server
$(go env GOPATH)/bin/bazelisk run //:lint 2>&1
echo "Bazel Lint Done"
