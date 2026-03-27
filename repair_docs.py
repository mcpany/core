import sys
import re

def repair(filename):
    with open(filename, 'r') as f:
        content = f.read()

    # Split into lines
    lines = content.split('\n')
    new_lines = []

    i = 0
    while i < len(lines):
        current = lines[i]

        # Check if the next line is exactly the same
        if i + 1 < len(lines) and current == lines[i+1] and current.strip():
            new_lines.append(current)
            i += 2
            continue

        # Check for header doubled parts (e.g., "# Market Sync: [2026-06-18]\n# Market Sync: [2026-06-18]**Status:** Complete")
        if current.startswith('#') and i + 1 < len(lines) and lines[i+1].startswith(current):
            # This looks like the case where a line was mashed
            # Use the more complete line
            new_lines.append(lines[i+1])
            i += 2
            continue

        new_lines.append(current)
        i += 1

    # Join and then do some regex for inline mashes
    result = '\n'.join(new_lines)

    # Example: "Complete**Status:**" -> "Complete\n**Status:**"
    # But more likely: "**Status:** Complete\n**Status:** Complete**Objective:**"
    # The previous dedupe attempt might have left some messy states.

    with open(filename, 'w') as f:
        f.write(result)

for arg in sys.argv[1:]:
    repair(arg)
