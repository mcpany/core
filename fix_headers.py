import re

def fix_header_length(filepath):
    with open(filepath, 'r') as f:
        lines = f.readlines()

    new_lines = []
    for line in lines:
        if line.startswith('### Focus:') and len(line) > 80:
            # We can't really wrap headers easily without breaking them,
            # but MD013 usually ignores headers if configured.
            # However, if CI fails, I should try to shorten them or wrap if possible.
            # Let's try to just split the Focus line into two lines if it's too long.
            # Wait, headers cannot be multiline.
            # I will try to use a shorter title for these historical headers.
            content = line.replace('### Focus: ', '')
            # If it's too long, maybe just keep it but I'll see if I can shorten.
            # Actually, I'll just leave them for now unless markdownlint complains.
            # MD013 often ignores headers.
            new_lines.append(line)
        else:
            new_lines.append(line)

    with open(filepath, 'w') as f:
        f.writelines(new_lines)

fix_header_length('docs/02_strategic_vision.md')
