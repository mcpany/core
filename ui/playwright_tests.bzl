# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

load("@rules_shell//shell:sh_test.bzl", "sh_test")

def _playwright_target_name(spec):
    return "playwright_" + spec.replace("/", "_").replace(".", "_").replace("-", "_")

def _playwright_target_data(spec):
    return [
        ":build",  # Pre-built Vite app (dist dir)
        ":node_modules",
        ":playwright_cli",
        ":playwright_runtime_srcs",
        ":playwright_support_srcs",
        spec,
        "//proto:ts_proto",
        "//server:config.minimal.yaml",
        "//server/cmd/server",
        "//server/tests/integration/cmd/mocks/http_echo_server",
    ]

def define_playwright_tests():
    targets = []
    for spec in sorted(native.glob(["tests/**/*.spec.ts"])):
        name = _playwright_target_name(spec)
        sh_test(
            name = name,
            srcs = ["playwright_test.sh"],
            args = [spec],
            data = _playwright_target_data(spec),
            size = "large",
            timeout = "long",
            tags = [
                "integration",
                "no-remote-exec",
            ],
        )
        targets.append(":" + name)

    native.test_suite(
        name = "playwright",
        tests = targets,
    )
