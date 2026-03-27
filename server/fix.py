import re

with open('pkg/app/api.go', 'r') as f:
    content = f.read()

# Instead of checking and ignoring `_ = w.Write` and adding `//nolint`, we should literally just do what the test failures ask for.
# But `//nolint:errcheck` IS a valid way to suppress it. The problem is the circle ci environment might be checking ALL `_ = ` assignments if `w.Write` is not explicitly disabled.
# Let's remove `_ = ` completely and just add `//nolint:errcheck` directly to the call.

lines = content.split('\n')
for i, line in enumerate(lines):
    if '//nolint:errcheck' in line:
        # replace `_, _ = w.Write` with `_, _ = w.Write` -> no wait, let's replace `_ = ` with nothing
        # Wait, `_, _ = w.Write` is an assignment. If we do `w.Write`, the lint rule `errcheck` requires `//nolint:errcheck`.
        # Let's change `_, _ = w.Write` to `w.Write`

        stripped = line.strip()
        indent = line[:len(line) - len(stripped)]

        if stripped.startswith('_, _ = '):
            # `_, _ = w.Write`
            expr = stripped[7:].split(' //nolint')[0].strip()
            lines[i] = f'{indent}{expr} //nolint:errcheck'
        elif stripped.startswith('_ = '):
            expr = stripped[4:].split(' //nolint')[0].strip()
            lines[i] = f'{indent}{expr} //nolint:errcheck'

        # also for defer func()
        if 'defer func() { _ = resp.Body.Close() }()' in lines[i]:
            lines[i] = lines[i].replace('_ = resp.Body.Close()', 'resp.Body.Close()')

with open('pkg/app/api.go', 'w') as f:
    f.write('\n'.join(lines))
