import re
import os
with open('scripts/lint.sh', 'r') as f:
    content = f.read()

# Oh, the script runs `golangci-lint` but ONLY if it finds it in bazel runfiles.
# wait, my `make lint` output literally says:
# Warning: golangci-lint not found (skipping Go linting).
# To enable, add a :golangci_lint_bin data dep or run 'make prepare'.
# It ALWAYS skipped golangci-lint in CI and local unless I ran the binary directly!
