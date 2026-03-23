import re

with open("server/pkg/storage/postgres/store_load_test.go", "r") as f:
    content = f.read()

content = content.replace("\tt.Parallel()\n", "")

with open("server/pkg/storage/postgres/store_load_test.go", "w") as f:
    f.write(content)
