#!/bin/bash
if [ -n "$BUILD_WORKSPACE_DIRECTORY" ]; then
  cd "$BUILD_WORKSPACE_DIRECTORY"
fi
exec ./bazel-bin/server/cmd/server/server_/server "$@"
