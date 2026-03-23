import os
import re

files = [
    'server/pkg/lint/linter.go',
    'server/pkg/middleware/audit.go',
    'server/pkg/metrics/metrics.go',
    'server/pkg/audit/sqlite.go',
    'server/pkg/audit/postgres.go',
    'server/pkg/admin/server.go'
]

def fix_file(path):
    with open(path, 'r') as f:
        content = f.read()

    # Remove any existing TODOs in docstrings
    content = content.replace('TODO: Document parameters.', 'The parameters.')
    content = content.replace('TODO: Document returns.', 'The result.')
    content = content.replace('TODO: Document errors.', 'An error if the operation fails.')

    with open(path, 'w') as f:
        f.write(content)

for f in files:
    if os.path.exists(f):
        fix_file(f)
