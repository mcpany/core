# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

"""Module extension for downloading pre-built golangci-lint binaries."""

_GOLANGCI_LINT_VERSION = "1.64.5"

# SHA256 checksums from:
# https://github.com/golangci/golangci-lint/releases/download/v1.64.5/golangci-lint-1.64.5-checksums.txt
_PLATFORMS = {
    "linux_amd64": {
        "sha256": "e6bd399a0479c5fd846dcf9f3990d20448b4f0d1e5027d82348eab9f80f7ac71",
        "os_arch": "linux-amd64",
        "strip_prefix": "golangci-lint-1.64.5-linux-amd64",
    },
    "linux_arm64": {
        "sha256": "59df27f9a82e461b00597c5f6d96c6a46bfdb4b7cddd9341502641d3d874a65a",
        "os_arch": "linux-arm64",
        "strip_prefix": "golangci-lint-1.64.5-linux-arm64",
    },
    "darwin_amd64": {
        "sha256": "7681c3e919491030558ef39b6ccaf49be1b3d19de611d30c02aec828dad822c1",
        "os_arch": "darwin-amd64",
        "strip_prefix": "golangci-lint-1.64.5-darwin-amd64",
    },
    "darwin_arm64": {
        "sha256": "8c4f11ef3a22d610dd5836a09c98e944b405624f932f20c7e72ae78abc552311",
        "os_arch": "darwin-arm64",
        "strip_prefix": "golangci-lint-1.64.5-darwin-arm64",
    },
}

def _golangci_lint_repo_impl(rctx):
    platform = rctx.attr.platform
    version = rctx.attr.version
    info = _PLATFORMS[platform]

    rctx.download_and_extract(
        url = "https://github.com/golangci/golangci-lint/releases/download/v{version}/golangci-lint-{version}-{os_arch}.tar.gz".format(
            version = version,
            os_arch = info["os_arch"],
        ),
        sha256 = info["sha256"],
        stripPrefix = info["strip_prefix"],
    )

    rctx.file("BUILD.bazel", """
package(default_visibility = ["//visibility:public"])

exports_files(["golangci-lint"])
""")

_golangci_lint_repo = repository_rule(
    implementation = _golangci_lint_repo_impl,
    attrs = {
        "platform": attr.string(mandatory = True),
        "version": attr.string(mandatory = True),
    },
)

def _golangci_lint_ext(mctx):
    for platform in _PLATFORMS:
        _golangci_lint_repo(
            name = "golangci_lint_{}".format(platform),
            platform = platform,
            version = _GOLANGCI_LINT_VERSION,
        )

golangci_lint = module_extension(
    implementation = _golangci_lint_ext,
)
