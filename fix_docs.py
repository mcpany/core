import os
import re

files = ["docs/02_strategic_vision.md", "docs/03_feature_inventory.md", "server/roadmap.md", "ui/roadmap.md"]

def wrap_line(line, width=80):
    if len(line) <= width:
        return [line]

    indent_match = re.match(r'^(\s*[-\*]?\s*)', line)
    indent = indent_match.group(1) if indent_match else ""

    # Special handling for bullet points: we want subsequent lines to be indented more
    bullet_match = re.match(r'^(\s*[-\*]\s+)', line)
    if bullet_match:
        bullet_indent = bullet_match.group(1)
        subsequent_indent = " " * len(bullet_indent)
    else:
        subsequent_indent = indent

    content = line[len(indent):]

    words = content.split(' ')
    lines = []
    curr = indent

    for i, w in enumerate(words):
        if not w and i != len(words)-1: continue
        if len(curr) + len(w) > width:
            lines.append(curr.rstrip())
            curr = subsequent_indent + w + " "
        else:
            curr += w + " "
    lines.append(curr.rstrip())
    return lines

for filepath in files:
    if not os.path.exists(filepath):
        continue
    with open(filepath, 'r', encoding='utf-8') as f:
        lines = f.readlines()

    new_lines = []
    for line in lines:
        line = line.rstrip('\n')
        # Clean non-ASCII
        line = line.replace('\u2014', '--').replace('\u2013', '-').replace('\u201c', '"').replace('\u201d', '"').replace('\u2018', "'").replace('\u2019', "'")

        if len(line) > 80:
            new_lines.extend(wrap_line(line))
        else:
            new_lines.append(line)

    with open(filepath, 'w', encoding='ascii', errors='replace') as f:
        f.write('\n'.join(new_lines) + '\n')
