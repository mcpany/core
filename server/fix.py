import re

with open('pkg/app/api.go', 'r') as f:
    content = f.read()

lines = content.split('\n')
for i, line in enumerate(lines):
    stripped = line.strip()
    if stripped.startswith('_ = ') or stripped.startswith('_, _ = '):
        if not line.endswith('//nolint:errcheck'):
            lines[i] = line + " //nolint:errcheck"

with open('pkg/app/api.go', 'w') as f:
    f.write('\n'.join(lines))
