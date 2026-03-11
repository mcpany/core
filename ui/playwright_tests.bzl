# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

load("@rules_shell//shell:sh_test.bzl", "sh_test")

def _playwright_target_name(spec):
    return "playwright_" + spec.replace("/", "_").replace(".", "_").replace("-", "_")

def _playwright_specs():
    return sorted(native.glob(["tests/**/*.spec.ts"]))

def _playwright_target_data(specs = []):
    return [
        ":next_cli",
        ":node_modules",
        ":playwright_cli",
        ":playwright_runtime_srcs",
        ":playwright_support_srcs",
        "//proto:ts_proto",
        "//server:config.minimal.yaml",
        "//server/cmd/server",
        "//server/tests/integration/cmd/mocks/http_echo_server",
    ] + specs

def define_playwright_tests():
    specs = _playwright_specs()
    sh_test(
        name = "playwright",
        srcs = ["playwright_test.sh"],
        data = _playwright_target_data(specs),
        size = "large",
        timeout = "eternal",
        tags = ["integration"],
    )
