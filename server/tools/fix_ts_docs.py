import os
import re

"""
This script automatically generates or updates JSDoc comments for React components
in TypeScript (.tsx) files. It attempts to infer prop descriptions based on names.
"""

def get_prop_desc(name):
    """
    Generates a description for a prop based on its name.

    Args:
        name: The name of the prop.

    Returns:
        A description string.
    """
    name_lower = name.lower()
    if name_lower == 'id': return "The unique identifier."
    if name_lower == 'name': return "The name."
    if name_lower == 'title': return "The title."
    if 'id' in name_lower: return f"The unique identifier for {name.replace('Id', '').replace('id', '').strip()}."
    if 'name' in name_lower: return f"The name of the {name.replace('Name', '').replace('name', '').strip()}."
    if 'children' in name_lower: return "The child components."
    if 'classname' in name_lower: return "Additional CSS classes."
    if 'open' in name_lower: return "Whether the component is open."
    if 'onchange' in name_lower: return "Callback function when value changes."
    if 'value' in name_lower: return "The current value."
    if 'schema' in name_lower: return "The schema definition."
    if 'required' in name_lower: return "Whether the field is required."
    if 'depth' in name_lower: return "The nesting depth."
    if 'auth' in name_lower: return "The authentication configuration."
    if 'type' in name_lower: return "The type definition."
    if 'status' in name_lower: return "The current status."
    if 'data' in name_lower: return "The data to display."
    if 'error' in name_lower: return "The error message or object."
    if 'loading' in name_lower: return "Whether data is loading."
    if 'config' in name_lower: return "The configuration object."
    return f"The {name} property."

def process_file(filepath):
    """
    Scans a file for React components and adds/updates docstrings.

    Args:
        filepath: The path to the file to process.
    """
    with open(filepath, 'r') as f:
        content = f.read()

    lines = content.split('\n')
    new_lines = []
    i = 0
    modified = False

    # regex for function/const definition (exported or not)
    # match widely, but only process if it looks like a component (Uppercase)
    def_pattern = re.compile(r'^\s*(export\s+)?(default\s+)?(function|const)\s+([a-zA-Z0-9_]+)')

    while i < len(lines):
        line = lines[i]
        match = def_pattern.match(line)

        if match:
            # Found a definition. Check for existing docstring.
            has_doc = False
            is_generic = False

            # Check previous lines
            j = i - 1
            while j >= 0:
                prev = lines[j].strip()
                if not prev:
                    j -= 1
                    continue
                if prev.startswith('@'): # decorators
                    j -= 1
                    continue
                if prev.endswith('*/'):
                    has_doc = True
                    # Check if it is the generic one we want to replace
                    # We check if it contains "@returns The rendered component." and NO @param

                    # Read back until /**
                    doc_content = ""
                    k = j
                    while k >= 0:
                        doc_content = lines[k] + "\n" + doc_content
                        if lines[k].strip() == '/**':
                            break
                        k -= 1

                    if ("@returns The rendered component." in doc_content and "@param" not in doc_content) or \
                       ("The name of the ." in doc_content):
                        is_generic = True

                break

            comp_name = match.group(4)

            # We only care about components (usually Capitalized)
            if comp_name and comp_name[0].isupper():

                # Parse params
                # Read lines until ')'
                func_sig = ""
                k = i
                while k < len(lines):
                    func_sig += lines[k]
                    if ')' in lines[k] or '=>' in lines[k]: # arrow function might use =>
                        # For const Name = (...) =>, checking ')' is usually enough for args
                        if ')' in lines[k]:
                            break
                    k += 1

                # Extract props
                # Case 1: React.forwardRef<T, P>(({ props }, ref) => ...
                # Case 2: function Name({ props }) ...

                params = []

                # Try to find object pattern destructuring inside function signature
                # We look for ({ ... }) pattern which is typical for props
                param_match = re.search(r'\(\s*({[^}]+})', func_sig, re.DOTALL)

                if param_match:
                    params_str = param_match.group(1)
                    params_str = params_str.replace('{', '').replace('}', '').replace('\n', '')
                    params_str = re.sub(r'//.*', '', params_str)
                    raw_params = [p.split('=')[0].split(':')[0].strip() for p in params_str.split(',')]
                    params = [p for p in raw_params if p and '...' not in p]

                if params or "React.forwardRef" in func_sig:
                    # Construct new docstring
                    new_doc = [
                        '/**',
                        f' * {comp_name} component.',
                        ' * @param props - The component props.'
                    ]
                    for p in params:
                        desc = get_prop_desc(p)
                        new_doc.append(f' * @param props.{p} - {desc}')
                    new_doc.append(' * @returns The rendered component.')
                    new_doc.append(' */')

                    if has_doc and is_generic:
                         # Remove old docstring from new_lines
                         idx = len(new_lines) - 1
                         while idx >= 0:
                             l = new_lines[idx].strip()
                             if not l or l.startswith('@'):
                                 idx -= 1
                                 continue
                             if l == '*/':
                                 end_idx = idx
                                 # Find start
                                 while idx >= 0:
                                     if new_lines[idx].strip() == '/**':
                                         start_idx = idx
                                         del new_lines[start_idx : end_idx+1]
                                         break
                                     idx -= 1
                                 break
                             break

                         insert_pos = len(new_lines)
                         while insert_pos > 0:
                             if new_lines[insert_pos-1].strip().startswith('@'):
                                 insert_pos -= 1
                             else:
                                 break

                         for l in new_doc:
                             new_lines.insert(insert_pos, l)
                             insert_pos += 1

                         modified = True
                         print(f"Upgraded docstring in {filepath} for {comp_name}")

                    elif not has_doc:
                         insert_pos = len(new_lines)
                         while insert_pos > 0:
                             if new_lines[insert_pos-1].strip().startswith('@'):
                                 insert_pos -= 1
                             else:
                                 break

                         for l in new_doc:
                             new_lines.insert(insert_pos, l)
                             insert_pos += 1

                         modified = True
                         print(f"Added docstring in {filepath} for {comp_name}")

        new_lines.append(line)
        i += 1

    if modified:
        with open(filepath, 'w') as f:
            f.write('\n'.join(new_lines))

def main():
    """
    Main function to walk the directory and process all .tsx files.
    """
    root_dir = 'ui/src'
    for dirpath, dirnames, filenames in os.walk(root_dir):
        for filename in filenames:
            if filename.endswith('.tsx'):
                process_file(os.path.join(dirpath, filename))

if __name__ == "__main__":
    main()
