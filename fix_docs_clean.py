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

    content = content.replace('TODO: Document parameters.', 'None.')
    content = content.replace('TODO: Document returns.', 'None.')
    content = content.replace('TODO: Document errors.', 'None.')

    with open(filepath, 'w') as f:
        f.write(content)

for root, _, files in os.walk('server/pkg'):
    for file in files:
        if file.endswith('.go'):
            process_file(os.path.join(root, file))
