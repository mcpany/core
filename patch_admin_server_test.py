import re

with open("server/pkg/admin/server_test.go", "r") as f:
    content = f.read()

content = content.replace('\t"log/slog"\n', '')
content = content.replace('\t"io"\n', '')

with open("server/pkg/admin/server_test.go", "w") as f:
    f.write(content)
