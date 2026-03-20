import os

replacements = {
    "TODO: Document parameters.": "Parameters for the operation.",
    "TODO: Document returns.": "The result of the operation.",
    "TODO: Document errors.": "An error if the operation fails."
}

def fix_docs(root):
    for root_dir, dirs, files in os.walk(root):
        for file in files:
            if file.endswith(".go"):
                path = os.path.join(root_dir, file)
                try:
                    with open(path, "r") as f:
                        content = f.read()

                    new_content = content
                    for old, new in replacements.items():
                        new_content = new_content.replace(old, new)

                    if new_content != content:
                        with open(path, "w") as f:
                            f.write(new_content)
                        print(f"Fixed {path}")
                except Exception as e:
                    print(f"Error fixing {path}: {e}")

if __name__ == '__main__':
    fix_docs("server/pkg")
    fix_docs("server/cmd")
    fix_docs("server/examples")
