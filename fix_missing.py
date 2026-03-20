import re
import os

def infer_param_doc(pname, ptype):
    pname_l = pname.lower()
    if pname_l == "ctx": return "The request context."
    if "err" in pname_l: return "An error object."
    if "id" in pname_l: return f"The identifier for the {pname.replace('ID', '').replace('Id', '')}."
    if "config" in pname_l: return "The configuration object."
    if "req" in pname_l: return "The request parameters."
    if "res" in pname_l: return "The response result."
    if "name" in pname_l: return "The name parameter."

    if ptype == "error": return "An error if the operation fails."
    if ptype.startswith("[]"): return "A list of items."
    if ptype.startswith("map"): return "A mapping of key-value pairs."
    if ptype == "bool": return "A boolean flag."
    if ptype == "string": return "A string value."
    if "int" in ptype: return "An integer value."

    return f"The {pname} parameter."

def infer_return_doc(rtype):
    rtype = rtype.strip()
    if rtype == "error": return "An error if the operation fails."
    if rtype == "bool": return "A boolean indicating success or status."
    if rtype == "string": return "The resulting string."
    if rtype.startswith("[]"): return "A list of results."
    if rtype.startswith("map"): return "A map of results."
    if rtype.startswith("*"): return f"A pointer to the {rtype[1:]} result."
    return f"The {rtype} result."

with open("missing.txt", "r") as f:
    missing = f.readlines()

files_to_update = {}
for line in missing:
    if not line.strip(): continue
    parts = line.strip().split(": ")
    filepath = parts[0]
    symbol_info = parts[1].split(" ")
    symbol_type = " ".join(symbol_info[:-1])
    symbol_name = symbol_info[-1]

    if filepath not in files_to_update:
        files_to_update[filepath] = []
    files_to_update[filepath].append((symbol_type, symbol_name))

for filepath, symbols in files_to_update.items():
    with open(filepath, 'r') as f:
        lines = f.read().split('\n')

    for symbol_type, symbol_name in symbols:
        target_pattern = None
        block_target_pattern = None

        if symbol_type == "func":
            target_pattern = re.compile(rf'^func\s+(?:\([^)]+\)\s+)?{symbol_name}\s*\(')
        elif symbol_type == "type":
            target_pattern = re.compile(rf'^type\s+{symbol_name}\s+')
        elif symbol_type == "interface method":
            if "." in symbol_name:
                symbol_name = symbol_name.split(".")[1]
            target_pattern = re.compile(rf'^\s+{symbol_name}\s*\(')
        elif symbol_type == "var/const":
            target_pattern = re.compile(rf'^(?:var|const)\s+{symbol_name}\s+')
            block_target_pattern = re.compile(rf'^\s+{symbol_name}\s*(?:=|type|[a-z])')

        for i, line in enumerate(lines):
            is_match = False
            is_block = False

            if target_pattern and target_pattern.match(line):
                is_match = True
                if symbol_type == "interface method":
                    is_block = True
            elif symbol_type == "var/const" and block_target_pattern and block_target_pattern.match(line):
                is_match = True
                is_block = True

            if is_match:
                j = i - 1
                has_summary = False
                while j >= 0 and lines[j].strip().startswith('//'):
                    if 'Summary:' in lines[j]:
                        has_summary = True
                        break
                    j -= 1

                if not has_summary:
                    indent_match = re.match(r'^(\s*)', line)
                    indent = indent_match.group(1) if indent_match else ""

                    if symbol_type == "func" or symbol_type == "interface method":
                        action = "Executes"
                        if symbol_name.startswith('Get') or symbol_name.startswith('Read') or symbol_name.startswith('List') or symbol_name.startswith('Describe') or symbol_name.startswith('Query'): action = "Retrieves"
                        elif symbol_name.startswith('Set') or symbol_name.startswith('Write') or symbol_name.startswith('Update') or symbol_name.startswith('Upsert') or symbol_name.startswith('Publish'): action = "Updates"
                        elif symbol_name.startswith('New') or symbol_name.startswith('Create') or symbol_name.startswith('Add') or symbol_name.startswith('Subscribe'): action = "Initializes"
                        elif symbol_name.startswith('Is') or symbol_name.startswith('Has'): action = "Checks"
                        elif symbol_name.startswith('Delete') or symbol_name.startswith('Remove'): action = "Deletes"

                        params = []
                        rets = []

                        sig_match = re.search(r'\((.*?)\)(?:\s+(.*?))?\{?$', line.split('{')[0].strip())
                        if sig_match:
                            params_str = sig_match.group(1)
                            rets_str = sig_match.group(2)

                            if params_str.strip():
                                for p in params_str.split(','):
                                    parts = p.strip().split()
                                    if len(parts) >= 2:
                                        params.append((parts[0], " ".join(parts[1:])))
                                    elif len(parts) == 1:
                                        params.append((parts[0], parts[0]))

                            if rets_str:
                                rets_str = rets_str.strip()
                                if rets_str.startswith('(') and rets_str.endswith(')'):
                                    rets_str = rets_str[1:-1]
                                for r in rets_str.split(','):
                                    parts = r.strip().split()
                                    if len(parts) >= 2:
                                        rets.append(parts[1])
                                    elif len(parts) == 1:
                                        rets.append(parts[0])

                        doc = f"{indent}// {symbol_name} ...\n"
                        doc += f"{indent}//\n"
                        doc += f"{indent}// Summary: {action} {symbol_name} operation.\n"
                        doc += f"{indent}//\n"
                        doc += f"{indent}// Parameters:\n"
                        if not params:
                            doc += f"{indent}//   - None.\n"
                        else:
                            for pname, ptype in params:
                                doc += f"{indent}//   - {pname}: {ptype}. {infer_param_doc(pname, ptype)}\n"
                        doc += f"{indent}//\n"
                        doc += f"{indent}// Returns:\n"
                        if not rets:
                            doc += f"{indent}//   - None.\n"
                        else:
                            for rtype in rets:
                                doc += f"{indent}//   - {rtype}: {infer_return_doc(rtype)}\n"
                        doc += f"{indent}//\n"
                        doc += f"{indent}// Errors:\n"
                        if "error" in rets or any("error" in t for _, t in params):
                            doc += f"{indent}//   - Returns error if the operation fails or is invalid.\n"
                        else:
                            doc += f"{indent}//   - None.\n"
                        doc += f"{indent}//\n"
                        doc += f"{indent}// Side Effects:\n"
                        doc += f"{indent}//   - None.\n"

                        lines[i] = doc + line

                    elif symbol_type == "type":
                        doc = f"{indent}// {symbol_name} ...\n"
                        doc += f"{indent}//\n"
                        doc += f"{indent}// Summary: Represents a {symbol_name}.\n"
                        lines[i] = doc + line

                    elif symbol_type == "var/const":
                        if is_block:
                            doc = f"{indent}// Summary: Defines {symbol_name}.\n"
                        else:
                            doc = f"{indent}// {symbol_name} ...\n"
                            doc += f"{indent}//\n"
                            doc += f"{indent}// Summary: Defines {symbol_name}.\n"
                        lines[i] = doc + line

    with open(filepath, 'w') as f:
        f.write('\n'.join(lines))
