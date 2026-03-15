cat << 'INNER_EOF' > BUILD.bazel
# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

load("@gazelle//:def.bzl", "gazelle")
load("@buildifier_prebuilt//:rules.bzl", "buildifier")

# gazelle:prefix github.com/mcpany/core
# gazelle:exclude k8s/
# gazelle:exclude build/
# gazelle:exclude ui/
# gazelle:exclude proto/third_party/
gazelle(name = "gazelle")

buildifier(
    name = "buildifier",
    lint_mode = "warn",
    mode = "fix",
)

sh_binary(
    name = "lint",
    srcs = ["scripts/lint.sh"],
    data = [
        "//server:go.mod",
        "//server:go.sum",
    ],
)

config_setting(
    name = "host_arm64",
    values = {"host_cpu": "aarch64"},
    visibility = ["//visibility:public"],
)
INNER_EOF
