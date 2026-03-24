import sys
import os

def check_go(filename):
    with open(filename, 'rb') as f:
        content = f.read()

    issues = []
    # Check for non-ASCII
    for i, b in enumerate(content):
        if b > 127:
            issues.append(f"Non-ASCII at byte {i}: {b}")
            break

    text = content.decode('utf-8', errors='replace')
    lines = text.splitlines()
    for i, line in enumerate(lines):
        # Trailing whitespace
        if line.endswith(' ') or line.endswith('\t'):
            issues.append(f"Trailing whitespace at line {i+1}")

    # Trailing newline
    if not text.endswith('\n'):
        issues.append("Missing trailing newline")
    elif text.endswith('\n\n'):
        issues.append("Multiple trailing newlines")

    return issues

files = [
    "server/pkg/app/server.go",
    "server/pkg/middleware/recursive_context.go",
    "server/pkg/middleware/recursive_context_test.go",
    "server/pkg/middleware/registry.go"
]

all_ok = True
for f in files:
    errs = check_go(f)
    if errs:
        all_ok = False
        print(f"--- {f} ---")
        for e in errs:
            print(e)
if all_ok:
    print("All Go files OK")
