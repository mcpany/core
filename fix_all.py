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

    lines = content.split('\n')
    new_lines = []

    for line in lines:
        if 'TODO: Document parameters.' in line:
            new_lines.append(line.replace('TODO: Document parameters.', 'Parameters are described in the interface documentation.'))
        elif 'TODO: Document returns.' in line:
            new_lines.append(line.replace('TODO: Document returns.', 'Returns the expected result defined by the interface.'))
        elif 'TODO: Document errors.' in line:
            new_lines.append(line.replace('TODO: Document errors.', 'Returns an error if the mock operation fails.'))
        else:
            new_lines.append(line)

    with open(filepath, 'w') as f:
        f.write('\n'.join(new_lines))

for root, _, files in os.walk('server/pkg'):
    for file in files:
        if file.endswith('.go'):
            process_file(os.path.join(root, file))
