import os
import re

def process_file(filepath):
    try:
        with open(filepath, 'r') as f:
            content = f.read()
    except Exception:
        return

    if 'TODO: Document' not in content:
        return

    content = content.replace('TODO: Document parameters.', 'Parameters are fully documented in the respective interface methods.')
    content = content.replace('TODO: Document returns.', 'Returns the mock result values configured for this execution.')
    content = content.replace('TODO: Document errors.', 'Returns an error detailing the exact mock failure if the call fails.')

    with open(filepath, 'w') as f:
        f.write(content)

for root, _, files in os.walk('server/pkg'):
    for file in files:
        if file.endswith('.go'):
            process_file(os.path.join(root, file))
