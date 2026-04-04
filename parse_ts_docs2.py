import os
import re

def main():
    missing_docs = 0
    total_funcs = 0
    for root, dirs, files in os.walk('ui/src'):
        for file in files:
            if file.endswith('.ts') or file.endswith('.tsx'):
                filepath = os.path.join(root, file)
                with open(filepath, 'r') as f:
                    content = f.read()

                # Match exported blocks
                # We can use regex to find 'export ' and then check the previous lines for block comments
                # Better yet, search for `@summary` or `Summary:` in the file
                # But we want to associate it with an export.

                # Let's write a simple TS parser:

                # Find all occurrences of "export "

                pass

main()
