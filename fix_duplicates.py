import re
from collections import defaultdict

def dedup_headers(filepath):
    with open(filepath, 'r') as f:
        lines = f.readlines()

    header_counts = defaultdict(int)
    new_lines = []

    for line in lines:
        match = re.match(r'^(#{1,6}\s+)(.*)', line)
        if match:
            prefix, content = match.groups()
            content = content.strip()
            header_counts[content] += 1
            if header_counts[content] > 1:
                # Add suffix to dedup
                new_line = f"{prefix}{content} ({header_counts[content]})\n"
                new_lines.append(new_line)
            else:
                new_lines.append(line)
        else:
            new_lines.append(line)

    with open(filepath, 'w') as f:
        f.writelines(new_lines)

files = [
    'docs/02_strategic_vision.md',
    'docs/03_feature_inventory.md',
    'server/roadmap.md'
]

for f in files:
    dedup_headers(f)
