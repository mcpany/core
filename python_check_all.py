import ast
import os
import sys

def check_file(filepath):
    missing = []
    with open(filepath, 'r') as f:
        try:
            tree = ast.parse(f.read(), filename=filepath)
        except Exception:
            return missing

    for node in tree.body:
        if isinstance(node, (ast.FunctionDef, ast.ClassDef, ast.AsyncFunctionDef)):
            # Ignore private methods/classes
            if node.name.startswith('_'):
                continue

            doc = ast.get_docstring(node)
            if not doc:
                missing.append((node.lineno, node.name, "Missing docstring"))
            else:
                if "Summary:" not in doc:
                    missing.append((node.lineno, node.name, "Missing Summary:"))
                if "Parameters:" not in doc:
                    missing.append((node.lineno, node.name, "Missing Parameters:"))
                if "Returns:" not in doc:
                    missing.append((node.lineno, node.name, "Missing Returns:"))
                if "Throws/Errors:" not in doc:
                    missing.append((node.lineno, node.name, "Missing Throws/Errors:"))
    return missing

has_err = False
for root, dirs, files in os.walk('.'):
    if 'node_modules' in dirs: dirs.remove('node_modules')
    if 'vendor' in dirs: dirs.remove('vendor')
    for file in files:
        if file.endswith('.py'):
            if "site-packages" in root or ".venv" in root or "patch_" in file or "fix_" in file:
                continue
            path = os.path.join(root, file)
            missing = check_file(path)
            for lineno, name, err in missing:
                print(f"{path}:{lineno}: {name} - {err}")
                has_err = True

if has_err:
    sys.exit(1)
