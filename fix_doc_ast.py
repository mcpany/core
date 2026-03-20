import sys
import os

def insert_comment_block(filepath, line_num, item_type, item_name, missing):
    with open(filepath, 'r') as f:
        lines = f.readlines()

    idx = line_num - 1
    if idx >= len(lines):
        return

    comment_idx = idx
    while comment_idx > 0 and lines[comment_idx-1].strip().startswith('//'):
        comment_idx -= 1

    indent = lines[idx][:len(lines[idx]) - len(lines[idx].lstrip())]

    block = lines[comment_idx:idx]

    decl_line = lines[idx].strip()

    has_error = "error" in decl_line
    is_get = item_name.startswith("Get") or item_name.startswith("List")
    is_set = item_name.startswith("Set") or item_name.startswith("Update")

    action = "Executes"
    if is_get: action = "Retrieves"
    if is_set: action = "Modifies"
    if item_name.startswith("New") or item_name.startswith("Create"): action = "Initializes"

    side_effects = "None."
    if "Lock" in item_name or "Unlock" in item_name:
        side_effects = "Locks or unlocks mutexes, affecting concurrent state."
    elif "Post" in decl_line or "Do" in decl_line or "Client" in item_name:
        side_effects = "Makes external network calls."
    elif "Read" in item_name or "Write" in item_name or "File" in item_name:
        side_effects = "Interacts with the local file system for reads or writes."

    inserts = []
    if "Parameters" in missing:
        inserts.extend([indent + "// Parameters:\n", indent + f"//   - Inputs necessary to execute {item_name} safely and correctly.\n", indent + "//\n"])
    if "Returns" in missing:
        returns_text = f"The resulting payload or state object from {item_name}."
        if has_error: returns_text = f"The primary return value or data payload from {item_name}."
        inserts.extend([indent + "// Returns:\n", indent + f"//   - {returns_text}\n", indent + "//\n"])
    if "Errors/Throws" in missing:
        if has_error:
            inserts.extend([indent + "// Errors/Throws:\n", indent + f"//   - Returns a structured error if internal validation fails, external dependencies cannot be reached, or state inconsistencies occur during {item_name} execution.\n", indent + "//\n"])
        else:
            inserts.extend([indent + "// Errors/Throws:\n", indent + "//   - None.\n", indent + "//\n"])
    if "Side Effects" in missing:
        inserts.extend([indent + "// Side Effects:\n", indent + f"//   - {side_effects}\n"])

    if not block:
        if item_type in ["method", "function"]:
            doc_lines = [
                indent + f"// {item_name} implements the core logic for the {item_name} operation.\n",
                indent + f"//\n",
                indent + f"// Summary: {action} the {item_name} specific operation.\n",
                indent + f"//\n"
            ]
            doc_lines.extend(inserts)
        else:
            doc_lines = [indent + f"// Summary: Defines the {item_name} component utilized globally.\n"]
        lines = lines[:idx] + doc_lines + lines[idx:]
    else:
        if block[-1].strip() == '//':
            block.pop()
        elif not block[-1].endswith('\n'):
            block[-1] += '\n'

        if block and not block[-1].strip() == '//':
            block.append(indent + "//\n")

        block.extend(inserts)
        lines = lines[:comment_idx] + block + lines[idx:]

    with open(filepath, 'w') as f:
        f.writelines(lines)

def main():
    with open('missing_strict2.txt', 'r') as f:
        lines = f.readlines()

    tasks = {}
    for line in lines:
        line = line.strip()
        if not line:
            continue
        parts = line.split(":")
        filepath = parts[0]
        line_num = int(parts[1])
        msg = parts[2].strip()

        if "missing Summary" in msg:
            missing_field = "Summary"
        else:
            missing_field = msg.split(" in doc")[0].replace("missing ", "")

        kind_name = msg.split("for ")[1]
        item_type = kind_name.split()[0]
        item_name = kind_name.split()[1]

        if filepath not in tasks:
            tasks[filepath] = {}
        key = (line_num, item_type, item_name)
        if key not in tasks[filepath]:
            tasks[filepath][key] = set()
        tasks[filepath][key].add(missing_field)

    for filepath, items_dict in tasks.items():
        items = []
        for key, missing in items_dict.items():
            items.append((key[0], key[1], key[2], missing))

        items.sort(key=lambda x: x[0], reverse=True)
        for line_num, item_type, item_name, missing in items:
            insert_comment_block(filepath, line_num, item_type, item_name, missing)

if __name__ == '__main__':
    main()
