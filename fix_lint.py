import os

files = [
    "docs/research/market-sync-2026-06-05.md",
    "docs/02_strategic_vision.md",
    "docs/03_feature_inventory.md",
    "docs/features/design-scpa.md",
    "server/roadmap.md",
    "ui/roadmap.md"
]

for filepath in files:
    if not os.path.exists(filepath):
        continue
    with open(filepath, 'r') as f:
        lines = f.readlines()

    new_lines = []
    for line in lines:
        # Remove trailing whitespace
        new_lines.append(line.rstrip() + '\n')

    # Ensure newline at end of file
    content = "".join(new_lines)
    if not content.endswith('\n'):
        content += '\n'

    with open(filepath, 'w') as f:
        f.write(content)
    print(f"Fixed {filepath}")
