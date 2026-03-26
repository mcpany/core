import re
import textwrap

def fix_content(content):
    # Ensure blank lines around headings
    content = re.sub(r'([^\n])\n(#{1,6} )', r'\1\n\n\2', content)
    content = re.sub(r'(#{1,6} [^\n]+)\n([^\n])', r'\1\n\n\2', content)

    # Ensure blank lines around lists (rough approximation)
    content = re.sub(r'([^\n])\n([-*] )', r'\1\n\n\2', content)
    content = re.sub(r'([-*] [^\n]+)\n([^-*\n ])', r'\1\n\n\2', content)

    lines = content.splitlines()
    fixed_lines = []
    in_code_block = False

    for line in lines:
        if line.startswith('```'):
            in_code_block = not in_code_block
            fixed_lines.append(line)
            continue

        if in_code_block:
            fixed_lines.append(line)
            continue

        if len(line) <= 80:
            fixed_lines.append(line)
            continue

        # Don't wrap headers
        if line.startswith('#'):
            fixed_lines.append(line)
            continue

        # Wrap long lines
        if line.startswith(('- ', '* ', '  ', '> ')):
            # Determine prefix
            prefix = ""
            if line.startswith(('- [ ] ', '- [x] ')):
                prefix = line[:6]
                text = line[6:]
                subsequent_indent = '      '
            elif line.startswith(('- ', '* ')):
                prefix = line[:2]
                text = line[2:]
                subsequent_indent = '  '
            else:
                # Just indented or quoted
                m = re.match(r'^(\s+|> )', line)
                prefix = m.group(1)
                text = line[len(prefix):]
                subsequent_indent = prefix

            wrapped = textwrap.fill(text, width=80, initial_indent=prefix, subsequent_indent=subsequent_indent, break_long_words=False)
            fixed_lines.extend(wrapped.splitlines())
        else:
            wrapped = textwrap.fill(line, width=80, break_long_words=False)
            fixed_lines.extend(wrapped.splitlines())

    return '\n'.join(fixed_lines)

files = [
    'docs/02_strategic_vision.md',
    'docs/03_feature_inventory.md',
    'server/roadmap.md',
    'ui/roadmap.md',
    'docs/features/design-csad-hub.md'
]

for fpath in files:
    with open(fpath, 'r') as f:
        original = f.read()

    # Before fixing, let's make sure we are not losing lines
    orig_line_count = len(original.splitlines())

    fixed = fix_content(original)

    # Simple check: fixed should have at least as many words/chars roughly
    if len(fixed) < len(original) * 0.8:
         print(f"Warning: significant size reduction in {fpath}")

    with open(fpath, 'w') as f:
        f.write(fixed + '\n')
