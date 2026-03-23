import os

directory = "ui/tests/e2e"

def remove_page_route_blocks(filepath):
    with open(filepath, "r") as f:
        content = f.read()

    new_content = ""
    i = 0
    while i < len(content):
        # Find start of `await page.route` or `await page.routeWebSocket`
        route_idx = content.find("await page.route", i)
        if route_idx == -1:
            new_content += content[i:]
            break

        # We found a route block.
        # Add everything before it.
        new_content += content[i:route_idx]

        # Now parse forward to find the matching ')' of the `page.route(...)` call.
        j = route_idx + len("await page.route")
        paren_count = 0
        started = False

        while j < len(content):
            if content[j] == '(':
                paren_count += 1
                started = True
            elif content[j] == ')':
                paren_count -= 1

            j += 1
            if started and paren_count == 0:
                break

        # Skip trailing semicolon and newline
        while j < len(content) and content[j] in [';', ' ', '\t']:
            j += 1
        if j < len(content) and content[j] == '\n':
            j += 1

        i = j

    # Also clean up empty lines that might have been left
    lines = new_content.split('\n')
    cleaned_lines = []
    empty_count = 0
    for line in lines:
        if line.strip() == "":
            empty_count += 1
            if empty_count <= 1:
                cleaned_lines.append(line)
        else:
            empty_count = 0
            cleaned_lines.append(line)

    with open(filepath, "w") as f:
        f.write("\n".join(cleaned_lines))

for filename in os.listdir(directory):
    if filename.endswith(".spec.ts"):
        filepath = os.path.join(directory, filename)
        remove_page_route_blocks(filepath)
