import os
import re

def process_file(filepath):
    with open(filepath, 'r') as f:
        content = f.read()

    lines = content.split('\n')
    new_lines = []
    i = 0

    while i < len(lines):
        line = lines[i]

        # Match exported function, type, var, const declarations
        match = re.match(r'^(func|type|var|const)\s+([A-Z][a-zA-Z0-9_]*)', line)
        method_match = re.match(r'^func\s+\([^)]+\)\s+([A-Z][a-zA-Z0-9_]*)', line)

        if match or method_match:
            name = match.group(2) if match else method_match.group(1)
            kind = match.group(1) if match else "func"

            # Look backwards to find the docstring
            j = i - 1
            comments = []
            while j >= 0 and lines[j].strip().startswith('//'):
                comments.insert(0, lines[j].strip())
                j -= 1

            has_summary = any('Summary:' in c for c in comments)

            if not has_summary:
                if kind == "type":
                    summary = f"// Summary: {name} represents a structure for {name}."
                elif kind == "func":
                    summary = f"// Summary: {name} executes the {name} logic."
                elif kind == "var" or kind == "const":
                    summary = f"// Summary: {name} defines the {name} value."

                if comments:
                    # Modify existing docstring
                    # Remove old comments from new_lines
                    del new_lines[j + 1:]

                    for c in comments:
                        new_lines.append(c)

                    new_lines.append('//')
                    new_lines.append(summary)

                    if kind == 'func':
                        # extract params
                        params = []
                        returns = []
                        # primitive param extraction
                        if "(" in line and ")" in line:
                            param_str = line.split("(")[1].split(")")[0]
                            if param_str:
                                p_parts = param_str.split(",")
                                for p in p_parts:
                                    if len(p.split()) >= 2:
                                        params.append(f"//   - {p.split()[0].strip()} ({p.split()[-1].strip()}): The {p.split()[0].strip()} parameter.")

                        ret_str = line.split(")")[-1].strip()
                        if ret_str.startswith("("):
                            ret_str = ret_str[1:-1]
                            r_parts = ret_str.split(",")
                            for r in r_parts:
                                returns.append(f"//   - {r.strip()}: The {r.strip()} return value.")
                        elif ret_str and "{" not in ret_str:
                            returns.append(f"//   - {ret_str.strip()}: The {ret_str.strip()} return value.")
                        elif ret_str and "{" in ret_str:
                            ret_str = ret_str.split("{")[0].strip()
                            if ret_str:
                                returns.append(f"//   - {ret_str.strip()}: The {ret_str.strip()} return value.")

                        new_lines.append('//')
                        new_lines.append('// Parameters:')
                        if params:
                            for p in params:
                                new_lines.append(p)
                        else:
                            new_lines.append('//   - None.')
                        new_lines.append('//')
                        new_lines.append('// Returns:')
                        if returns:
                            for r in returns:
                                new_lines.append(r)
                        else:
                            new_lines.append('//   - None.')
                        new_lines.append('//')
                        new_lines.append('// Errors:')
                        if any("error" in r for r in returns):
                            new_lines.append('//   - Returns an error if the operation fails.')
                        else:
                            new_lines.append('//   - None.')

                    new_lines.append(line)
                else:
                    # Add new docstring
                    new_lines.append(f"// {name} ...")
                    new_lines.append('//')
                    new_lines.append(summary)
                    if kind == 'func':
                        # extract params
                        params = []
                        returns = []
                        # primitive param extraction
                        if "(" in line and ")" in line:
                            param_str = line.split("(")[1].split(")")[0]
                            if param_str:
                                p_parts = param_str.split(",")
                                for p in p_parts:
                                    if len(p.split()) >= 2:
                                        params.append(f"//   - {p.split()[0].strip()} ({p.split()[-1].strip()}): The {p.split()[0].strip()} parameter.")

                        ret_str = line.split(")")[-1].strip()
                        if ret_str.startswith("("):
                            ret_str = ret_str[1:-1]
                            r_parts = ret_str.split(",")
                            for r in r_parts:
                                returns.append(f"//   - {r.strip()}: The {r.strip()} return value.")
                        elif ret_str and "{" not in ret_str:
                            returns.append(f"//   - {ret_str.strip()}: The {ret_str.strip()} return value.")
                        elif ret_str and "{" in ret_str:
                            ret_str = ret_str.split("{")[0].strip()
                            if ret_str:
                                returns.append(f"//   - {ret_str.strip()}: The {ret_str.strip()} return value.")

                        new_lines.append('//')
                        new_lines.append('// Parameters:')
                        if params:
                            for p in params:
                                new_lines.append(p)
                        else:
                            new_lines.append('//   - None.')
                        new_lines.append('//')
                        new_lines.append('// Returns:')
                        if returns:
                            for r in returns:
                                new_lines.append(r)
                        else:
                            new_lines.append('//   - None.')
                        new_lines.append('//')
                        new_lines.append('// Errors:')
                        if any("error" in r for r in returns):
                            new_lines.append('//   - Returns an error if the operation fails.')
                        else:
                            new_lines.append('//   - None.')
                    new_lines.append(line)
            else:
                new_lines.append(line)
        else:
            new_lines.append(line)
        i += 1

    with open(filepath, 'w') as f:
        f.write('\n'.join(new_lines))

for root, _, files in os.walk('server/pkg'):
    for file in files:
        if file.endswith('.go') and not file.endswith('_test.go'):
            process_file(os.path.join(root, file))
