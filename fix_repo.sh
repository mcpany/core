#!/bin/bash
cat << 'MODULE' >> MODULE.bazel
use_repo(
    go_deps,
    "org_golang_google_genproto_googleapis_api",
)
MODULE
