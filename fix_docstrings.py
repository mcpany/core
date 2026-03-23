import re
import os

filepath = "server/pkg/lint/linter.go"

with open(filepath, 'r') as f:
    lines = f.readlines()

def fix_doc(doc_lines):
    # Standard sections: Summary, Parameters, Returns, Errors, Side Effects
    sections = {
        'Summary': '',
        'Parameters': [],
        'Returns': [],
        'Errors': [],
        'Side Effects': []
    }

    # Extract existing info
    curr_section = None
    for line in doc_lines:
        line = line.strip(' /').strip()
        if not line: continue

        if line.startswith('Summary:'):
            sections['Summary'] = line.replace('Summary:', '').strip()
            curr_section = 'Summary'
        elif line.startswith('Parameters:'):
            curr_section = 'Parameters'
        elif line.startswith('Returns:'):
            curr_section = 'Returns'
        elif line.startswith('Errors:'):
            curr_section = 'Errors'
        elif line.startswith('Side Effects:'):
            curr_section = 'Side Effects'
        elif curr_section in ['Parameters', 'Returns', 'Errors', 'Side Effects']:
            sections[curr_section].append(line)

    # Reconstruct
    new_doc = []
    if sections['Summary']:
        new_doc.append(f"// Summary: {sections['Summary']}")
        new_doc.append("//")

    if sections['Parameters']:
        new_doc.append("// Parameters:")
        for p in sections['Parameters']:
            new_doc.append(f"//   {p}")
    else:
        new_doc.append("// Parameters:")
        new_doc.append("//   - None.")

    new_doc.append("//")

    if sections['Returns']:
        new_doc.append("// Returns:")
        for r in sections['Returns']:
            new_doc.append(f"//   {r}")
    else:
        new_doc.append("// Returns:")
        new_doc.append("//   - None.")

    new_doc.append("//")

    if sections['Errors']:
        new_doc.append("// Errors:")
        for e in sections['Errors']:
            new_doc.append(f"//   {e}")
    else:
        new_doc.append("// Errors:")
        new_doc.append("//   - None.")

    new_doc.append("//")

    if sections['Side Effects']:
        new_doc.append("// Side Effects:")
        for s in sections['Side Effects']:
            new_doc.append(f"//   {s}")
    else:
        new_doc.append("// Side Effects:")
        new_doc.append("//   - None.")

    return new_doc

new_file_lines = []
i = 0
while i < len(lines):
    line = lines[i]
    if line.startswith('//') and not line.startswith('// Copyright') and not line.startswith('// SPDX') and not line.startswith('// Package'):
        # Collect doc block
        doc_block = []
        while i < len(lines) and lines[i].startswith('//'):
            doc_block.append(lines[i])
            i += 1

        # Check if it's followed by a func
        if i < len(lines) and lines[i].startswith('func'):
            # Fix it
            new_file_lines.extend([l + '\n' for l in fix_doc(doc_block)])
        else:
            new_file_lines.extend(doc_block)
    else:
        new_file_lines.append(line)
        i += 1

with open(filepath, 'w') as f:
    f.writelines(new_file_lines)
