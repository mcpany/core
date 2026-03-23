import os
import re

directory = "ui/tests/e2e"

def process_file(filepath):
    with open(filepath, "r") as f:
        content = f.read()

    # Find the bounds of 'test.beforeEach' or similar blocks where route mocks are usually placed.
    # A more robust regex approach: replace `await page.route(..., async route => { ... });`
    # We will use a stack to find matching braces and completely remove the page.route blocks.

    new_content = ""
    i = 0
    while i < len(content):
        # We also need to match page.routeWebSocket
        match = re.search(r"await page\.route(?:WebSocket)?\s*\(", content[i:])
        if match:
            start_index = i + match.start()
            # Append everything before the match
            new_content += content[i:start_index]

            # Now we need to parse ahead to find the matching closing brace/parenthesis
            # It's an async call: `await page.route(..., async route => { ... });`
            j = start_index + len(match.group(0))
            paren_count = 1
            brace_count = 0

            while j < len(content):
                if content[j] == '(':
                    paren_count += 1
                elif content[j] == ')':
                    paren_count -= 1
                elif content[j] == '{':
                    brace_count += 1
                elif content[j] == '}':
                    brace_count -= 1

                j += 1
                if paren_count == 0 and brace_count == 0:
                    # Found the end of the route block
                    # Skip the trailing semicolon and newline if present
                    if j < len(content) and content[j] == ';':
                        j += 1
                    if j < len(content) and content[j] == '\n':
                        j += 1
                    break

            i = j
        else:
            new_content += content[i:]
            break

    with open(filepath, "w") as f:
        f.write(new_content)

for filename in os.listdir(directory):
    if filename.endswith(".spec.ts"):
        filepath = os.path.join(directory, filename)
        process_file(filepath)
