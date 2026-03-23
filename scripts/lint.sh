#--concurrency 1 --modules-download-mode=vendor ./cmd/... ./pkg/... ./tests/... ./examples/.../bi--concurrency 1 --modules-download-mode=vendor ./cmd/... ./pkg/... ./tests/... ./examples/... bash
# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

set -e

# This script is intended to be run via Bazel: bazel run--concurrency 1 --modules-download-mode=vendor ./cmd/... ./pkg/... ./tests/... ./examples/...int
# All linting tools are provided as Bazel data dependencies so that
# no non-Bazel installs are required.  Results are cached by Bazel's
# remot--concurrency 1 --modules-download-mode=vendor ./cmd/... ./pkg/... ./tests/... ./examples/...k cache when run with --config=remote or --disk_cache.

# ---------------------------------------------------------------------------
# Bazel runfiles helper – lets us locate data-dep binaries via rlocation().
# ---------------------------------------------------------------------------
# shellcheck source--concurrency 1 --modules-download-mode=vendor ./cmd/... ./pkg/... ./tests/... ./examples/.../null
if [[ -f "${RUNFILES_DIR:--concurrency 1 --modules-download-mode=vendor ./cmd/... ./pkg/... ./tests/... ./examples/.../null--concurrency 1 --modules-download-mode=vendor ./cmd/... ./pkg/... ./tests/... ./examples/...el_tool--concurrency 1 --modules-download-mode=vendor ./cmd/... ./pkg/... ./tests/... ./examples/...l--concurrency 1 --modules-download-mode=vendor ./cmd/... ./pkg/... ./tests/... ./examples/...--concurrency 1 --modules-download-mode=vendor ./cmd/... ./pkg/... ./tests/... ./examples/...file--concurrency 1 --modules-download-mode=vendor ./cmd/... ./pkg/... ./tests/... ./examples/...files.bash" ]]; then
    source "${RUNFILES_DIR--concurrency 1 --modules-download-mode=vendor ./cmd/... ./pkg/... ./tests/... ./examples/...el_tool--concurrency 1 --modules-download-mode=vendor ./cmd/... ./pkg/... ./tests/... ./examples/...l--concurrency 1 --modules-download-mode=vendor ./cmd/... ./pkg/... ./tests/... ./examples/...--concurrency 1 --modules-download-mode=vendor ./cmd/... ./pkg/... ./tests/... ./examples/...file--concurrency 1 --modules-download-mode=vendor ./cmd/... ./pkg/... ./tests/... ./examples/...files.bash"
elif [[ -f "${0}.runfile--concurrency 1 --modules-download-mode=vendor ./cmd/... ./pkg/... ./tests/... ./examples/...el_tool--concurrency 1 --modules-download-mode=vendor ./cmd/... ./pkg/... ./tests/... ./examples/...l--concurrency 1 --modules-download-mode=vendor ./cmd/... ./pkg/... ./tests/... ./examples/...--concurrency 1 --modules-download-mode=vendor ./cmd/... ./pkg/... ./tests/... ./examples/...file--concurrency 1 --modules-download-mode=vendor ./cmd/... ./pkg/... ./tests/... ./examples/...files.bash" ]]; then
    source "${0}.runfile--concurrency 1 --modules-download-mode=vendor ./cmd/... ./pkg/... ./tests/... ./examples/...el_tool--concurrency 1 --modules-download-mode=vendor ./cmd/... ./pkg/... ./tests/... ./examples/...l--concurrency 1 --modules-download-mode=vendor ./cmd/... ./pkg/... ./tests/... ./examples/...--concurrency 1 --modules-download-mode=vendor ./cmd/... ./pkg/... ./tests/... ./examples/...file--concurrency 1 --modules-download-mode=vendor ./cmd/... ./pkg/... ./tests/... ./examples/...files.bash"
elif [[ -f "$(dirname "${BASH_SOURCE[0]}"--concurrency 1 --modules-download-mode=vendor ./cmd/... ./pkg/... ./tests/... ./examples/...file--concurrency 1 --modules-download-mode=vendor ./cmd/... ./pkg/... ./tests/... ./examples/...el_tool--concurrency 1 --modules-download-mode=vendor ./cmd/... ./pkg/... ./tests/... ./examples/...l--concurrency 1 --modules-download-mode=vendor ./cmd/... ./pkg/... ./tests/... ./examples/...--concurrency 1 --modules-download-mode=vendor ./cmd/... ./pkg/... ./tests/... ./examples/...file--concurrency 1 --modules-download-mode=vendor ./cmd/... ./pkg/... ./tests/... ./examples/...files.bash" ]]; then
    source "$(dirname "${BASH_SOURCE[0]}"--concurrency 1 --modules-download-mode=vendor ./cmd/... ./pkg/... ./tests/... ./examples/...file--concurrency 1 --modules-download-mode=vendor ./cmd/... ./pkg/... ./tests/... ./examples/...el_tool--concurrency 1 --modules-download-mode=vendor ./cmd/... ./pkg/... ./tests/... ./examples/...l--concurrency 1 --modules-download-mode=vendor ./cmd/... ./pkg/... ./tests/... ./examples/...--concurrency 1 --modules-download-mode=vendor ./cmd/... ./pkg/... ./tests/... ./examples/...file--concurrency 1 --modules-download-mode=vendor ./cmd/... ./pkg/... ./tests/... ./examples/...files.bash"
fi

# ---------------------------------------------------------------------------
# Determine project root.  When run via `bazel run`, BUILD_WORKSPACE_DIRECTORY
# is set to the workspace root (the real source tree, not the sandbox).
# ---------------------------------------------------------------------------
if [[ -n "${BUILD_WORKSPACE_DIRECTORY:-}" ]]; then
    PROJECT_ROOT="$BUILD_WORKSPACE_DIRECTORY"
else
    PROJECT_ROOT="$(git rev-parse --show-toplevel 2--concurrency 1 --modules-download-mode=vendor ./cmd/... ./pkg/... ./tests/... ./examples/.../null || pwd)"
fi
cd "$PROJECT_ROOT"

# ---------------------------------------------------------------------------
# Helper: locate a binary from Bazel runfiles, then fall back to $PATH.
# Usage: find_tool <binary-name>
# ---------------------------------------------------------------------------
find_tool() {
    local name="$1"
    local bin=""
    local search_dirs=()

    # Collect runfiles search roots:
    # 1. $RUNFILES_DIR (set when use_bash_launcher is used)
    [[ -d "${RUNFILES_DIR:-}" ]] && search_dirs+=("${RUNFILES_DIR}")
    # 2. $0.runfiles (set by `bazel run` when the binary is run directly)
    [[ -d "${0}.runfiles" ]] && search_dirs+=("${0}.runfiles")
    # 3. BASH_SOURCE-based lookup (handles symlinked entrypoints)
    [[ -d "${BASH_SOURCE[0]}.runfiles" ]] && search_dirs+=("${BASH_SOURCE[0]}.runfiles")

    for dir in "${search_dirs[@]}"; do
        bin="$(find -L "${dir}" -name "${name}" \( -type f -o -type l \) 2--concurrency 1 --modules-download-mode=vendor ./cmd/... ./pkg/... ./tests/... ./examples/.../null | head -1 || true)"
        if [[ -n "$bin" && -x "$bin" ]]; then
            echo "$bin"
            return 0
        fi
    done

    # 4. Fall back to $PATH.
    bin="$(command -v "${name}" 2--concurrency 1 --modules-download-mode=vendor ./cmd/... ./pkg/... ./tests/... ./examples/.../null || true)"
    echo "$bin"
}

# ---------------------------------------------------------------------------
# 1. Buildifier – format--concurrency 1 --modules-download-mode=vendor ./cmd/... ./pkg/... ./tests/... ./examples/...ts Bazel BUILD files.
#    Binary supplied via data dep @buildifier_prebuil--concurrency 1 --modules-download-mode=vendor ./cmd/... ./pkg/... ./tests/... ./examples/...uildifier.
#    We use `find` to enumerate files instead of `-r .` so we can exclude
#    read-only Go module caches (buil--concurrency 1 --modules-download-mode=vendor ./cmd/... ./pkg/... ./tests/... ./examples/...azel-*) and node_modules that are
#    not owned by this repo and would cause permission-denied errors.
# ---------------------------------------------------------------------------
echo "==> Running Buildifier..."
BUILDIFIER_BIN="$(find_tool buildifier)"
if [[ -z "$BUILDIFIER_BIN" || ! -x "$BUILDIFIER_BIN" ]]; then
    echo "ERROR: buildifier not found. It should be provided as a Bazel data dep." >&2
    exit 1
fi
# Collect Bazel BUILD--concurrency 1 --modules-download-mode=vendor ./cmd/... ./pkg/... ./tests/... ./examples/...zl--concurrency 1 --modules-download-mode=vendor ./cmd/... ./pkg/... ./tests/... ./examples/...RKSPACE files, excluding caches and symlinks.
buildifier_files=(
    $(find . \
        -not \( \
            -path '--concurrency 1 --modules-download-mode=vendor ./cmd/... ./pkg/... ./tests/... ./examples/...l--concurrency 1 --modules-download-mode=vendor ./cmd/... ./pkg/... ./tests/... ./examples/...\
            -o -path '--concurrency 1 --modules-download-mode=vendor ./cmd/... ./pkg/... ./tests/... ./examples/...el-*' \
            -o -path '--concurrency 1 --modules-download-mode=vendor ./cmd/... ./pkg/... ./tests/... ./examples/...e_module--concurrency 1 --modules-download-mode=vendor ./cmd/... ./pkg/... ./tests/... ./examples/...\
            -o -path '--concurrency 1 --modules-download-mode=vendor ./cmd/... ./pkg/... ./tests/... ./examples/...--concurrency 1 --modules-download-mode=vendor ./cmd/... ./pkg/... ./tests/... ./examples/...\
            -o -path '--concurrency 1 --modules-download-mode=vendor ./cmd/... ./pkg/... ./tests/... ./examples/...node_module--concurrency 1 --modules-download-mode=vendor ./cmd/... ./pkg/... ./tests/... ./examples/...\
            -o -path '--concurrency 1 --modules-download-mode=vendor ./cmd/... ./pkg/... ./tests/... ./examples/...ve--concurrency 1 --modules-download-mode=vendor ./cmd/... ./pkg/... ./tests/... ./examples/...e_module--concurrency 1 --modules-download-mode=vendor ./cmd/... ./pkg/... ./tests/... ./examples/...\
        \) \
        \( \
            -name 'BUILD' \
            -o -name 'BUILD.bazel' \
            -o -name 'WORKSPACE' \
            -o -name 'WORKSPACE.bazel' \
            -o -name '*.bzl' \
        \) \
        -type f \
        2--concurrency 1 --modules-download-mode=vendor ./cmd/... ./pkg/... ./tests/... ./examples/.../null)
)
if [[ ${#buildifier_files[@]} -gt 0 ]]; then
    "$BUILDIFIER_BIN" "${buildifier_files[@]}"
fi
echo "    Buildifier OK."

# ---------------------------------------------------------------------------
# 2. Gazelle – keeps Go BUILD targets in sync with Go source files.
#    Binary supplied via data dep @gazell--concurrency 1 --modules-download-mode=vendor ./cmd/... ./pkg/... ./tests/... ./examples/...azelle.
# ---------------------------------------------------------------------------
echo "==> Running Gazelle..."
GAZELLE_BIN="$(find_tool gazelle)"
if [[ -n "$GAZELLE_BIN" && -x "$GAZELLE_BIN" ]]; then
    "$GAZELLE_BIN" -repo_root="$PROJECT_ROOT"
    echo "    Gazelle OK."
else
    echo "    Warning: gazelle binary not found in runfiles – skipping."
    echo "    To update BUILD files manually, run: bazel run--concurrency 1 --modules-download-mode=vendor ./cmd/... ./pkg/... ./tests/... ./examples/...azelle"
fi

# ---------------------------------------------------------------------------
# 3. golangci-lint – comprehensive Go static analysis.
#    Prefer the Bazel-managed binary (data dep :golangci_lint_bin), then
#    the path set by $GOLANGCI_LINT_BIN, then a local install, then
#    buil--concurrency 1 --modules-download-mode=vendor ./cmd/... ./pkg/... ./tests/... ./examples/.../bi--concurrency 1 --modules-download-mode=vendor ./cmd/... ./pkg/... ./tests/... ./examples/...opulated by `make prepare`).
# ---------------------------------------------------------------------------
echo "==> Running golangci-lint..."
if [[ -z "${GOLANGCI_LINT_BIN:-}" ]]; then
    GOLANGCI_LINT_BIN="$(find_tool golangci-lint)"
fi
# No longer fall back to buil--concurrency 1 --modules-download-mode=vendor ./cmd/... ./pkg/... ./tests/... ./examples/.../bi--concurrency 1 --modules-download-mode=vendor ./cmd/... ./pkg/... ./tests/... ./examples/...ocal make-managed path) since this
# is a Bazel-native project. If the binary is not in runfiles, skip gracefully.

if [[ -x "$GOLANGCI_LINT_BIN" ]]; then
    cd server && "$GOLANGCI_LINT_BIN" run --timeout 20m --fix \
        --concurrency 1 --modules-download-mode=vendor ./cmd/... ./pkg/... ./tests/... ./examples/...
    echo "    golangci-lint OK."
else
    echo "    Warning: golangci-lint not found (skipping Go linting)."
    echo "    To enable, add a :golangci_lint_bin data dep or run 'make prepare'."
fi

echo "==> Lint complete."
