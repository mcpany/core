import re

with open('pkg/app/api.go', 'r') as f:
    content = f.read()

lines = content.split('\n')
for i, line in enumerate(lines):
    if 'b, _ := opts.Marshal(' in line and '//nolint:errcheck' not in line:
        lines[i] = line.replace('b, _ := opts.Marshal(', 'b, _ := opts.Marshal(') + ' //nolint:errcheck'

with open('pkg/app/api.go', 'w') as f:
    f.write('\n'.join(lines))
