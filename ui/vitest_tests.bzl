# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

"""Defines per-file Vitest Bazel targets for better cacheability."""

def define_vitest_tests(vitest_bin):
    test_files = sorted(native.glob([
        "src/**/*.test.ts",
        "src/**/*.test.tsx",
        "tests/**/*.test.ts",
        "tests/**/*.test.tsx",
    ]))

    test_targets = []
    for test_file in test_files:
        target_name = "vitest_" + test_file.replace("/", "_").replace(".", "_").replace("-", "_")
        test_targets.append(":" + target_name)
        vitest_bin.vitest_test(
            name = target_name,
            args = [
                "run",
                test_file,
            ],
            chdir = native.package_name(),
            data = [
                ":node_modules",
                ":srcs",
                ":playwright_support_srcs",
                "//proto:ts_proto",
            ],
        )

    native.test_suite(
        name = "vitest",
        tests = test_targets,
    )
