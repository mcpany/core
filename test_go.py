import re
content = """// Test does something.
func Test() {
}
"""
matches = re.finditer(r'^(?:\/\/(?:[^\n]*)\n)*func (?:\([^\)]+\)\s+)?([A-Z][a-zA-Z0-9_]*)\(', content, re.MULTILINE)
for match in matches:
    print(match.group(0))
