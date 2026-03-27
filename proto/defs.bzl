# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

load("@aspect_rules_ts//ts:proto.bzl", "ts_proto_library")

def mcp_ts_proto_library(name, proto, deps = []):
    ts_proto_library(
        name = name,
        node_modules = "//ui:node_modules",
        proto = proto,
        gen_connect_es = True,
        copy_files = False,
        visibility = ["//visibility:public"],
    )
