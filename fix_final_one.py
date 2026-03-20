import re

with open("./server/docs/features/webhooks/examples/html_to_md/main.go", "r") as f:
    content = f.read()

content = content.replace("const StatusOK = 200", "// Summary: StatusOK indicates a successful webhook response.\nconst StatusOK = 200")

with open("./server/docs/features/webhooks/examples/html_to_md/main.go", "w") as f:
    f.write(content)
