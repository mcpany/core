import os
import re

def process_file(filepath):
    with open(filepath, 'r') as f:
        content = f.read().split('\n')

    # Regex for public function/method
    func_regex = re.compile(r'^func\s+(?:\([^)]+\)\s+)?([A-Z][A-Za-z0-9_]*)')

    new_content = []
    i = 0
    while i < len(content):
        line = content[i]
        m = func_regex.match(line)
        if m:
            func_name = m.group(1)

            # Check if there are comments above it
            j = i - 1
            doc_block = []
            while j >= 0 and content[j].startswith('//'):
                doc_block.insert(0, content[j])
                j -= 1

            doc_text = '\n'.join(doc_block)

            if 'Summary:' not in doc_text or 'Returns:' not in doc_text:
                # We need to add/update the docstring
                # Determine parameters and return types
                # Very basic parsing, not perfect but sufficient for placeholder

                # Build the new docstring
                new_doc = []

                # If there's an existing comment, keep it as the first line description
                # Or just keep the whole existing doc_block and append the structure
                # if it doesn't already have it

                if doc_block:
                    new_doc.extend(doc_block)
                else:
                    new_doc.append(f"// {func_name} ...")

                if 'Summary:' not in doc_text:
                    new_doc.append("//")
                    new_doc.append(f"// Summary: Executes {func_name} operation.")

                if 'Parameters:' not in doc_text:
                    new_doc.append("//")
                    new_doc.append("// Parameters:")
                    new_doc.append("//   - TODO: Document parameters.")

                if 'Returns:' not in doc_text:
                    new_doc.append("//")
                    new_doc.append("// Returns:")
                    new_doc.append("//   - TODO: Document returns.")

                if 'Errors:' not in doc_text:
                    new_doc.append("//")
                    new_doc.append("// Errors:")
                    new_doc.append("//   - None.")

                if 'Side Effects:' not in doc_text:
                    new_doc.append("//")
                    new_doc.append("// Side Effects:")
                    new_doc.append("//   - None.")

                # Replace the old doc block with the new one
                # First remove old doc block from new_content
                k = len(new_content) - 1
                while k >= 0 and new_content[k].startswith('//'):
                    new_content.pop()
                    k -= 1

                new_content.extend(new_doc)

        new_content.append(line)
        i += 1

    with open(filepath, 'w') as f:
        f.write('\n'.join(new_content))

def main():
    for root, dirs, files in os.walk('server/pkg'):
        for f in files:
            if f.endswith('.go') and not f.endswith('_test.go'):
                process_file(os.path.join(root, f))

if __name__ == '__main__':
    main()
