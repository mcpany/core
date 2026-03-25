with open("server/pkg/cli/cli.go", "r") as f:
    lines = f.readlines()

has_package = False
for l in lines:
    if l.startswith("package "):
        has_package = True

if not has_package:
    new_lines = ["package cli\n\n"] + lines
    with open("server/pkg/cli/cli.go", "w") as f:
        f.writelines(new_lines)
