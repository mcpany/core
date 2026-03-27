#!/bin/bash

# Ensure stretchr/mock is available for our test logic
sed -i 's/"@com_github_stretchr_testify\/\/require",/"@com_github_stretchr_testify\/\/require",\n        "@com_github_stretchr_testify\/\/mock",/g' server/pkg/logging/BUILD.bazel
