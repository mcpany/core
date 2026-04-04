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

                # Simple regex for exported functions and consts
                matches = re.finditer(r'(?:\/\*\*[\s\S]*?\*\/[\s\n]*)?export (?:const|function|class|interface|type)\s+([A-Z][a-zA-Z0-9_]*)', content, re.MULTILINE)
                for match in matches:
                    total_funcs += 1
                    block = match.group(0)
                    if "@summary" not in block.lower() and "summary" not in block.lower():
                        missing_docs += 1
                        print(f"{filepath}: {match.group(1)}")
    print(f"Missing docs in {missing_docs} out of {total_funcs} exported items in ui/src")

main()
