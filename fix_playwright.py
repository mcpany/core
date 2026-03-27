with open("ui/playwright_test.sh", "r") as f:
    lines = f.readlines()
for i, line in enumerate(lines):
    if 'escaped_spec="$(printf "%s" "$selected_spec"' in line:
        lines[i] = r"""  escaped_spec="$(printf '%s' "$selected_spec" | sed -e 's/[.[\*^$()+?{}|]/\\&/g')"
"""
with open("ui/playwright_test.sh", "w") as f:
    f.writelines(lines)
