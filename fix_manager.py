with open("server/pkg/webhooks/manager.go", "r") as f:
    lines = f.readlines()

new_lines = []
has_package = False
for l in lines:
    if l.startswith("package "):
        has_package = True
    new_lines.append(l)

if not has_package:
    new_lines.insert(0, "package webhooks\n\n")

with open("server/pkg/webhooks/manager.go", "w") as f:
    f.writelines(new_lines)
