import re

with open("server/pkg/storage/postgres/store_load_test.go", "r") as f:
    content = f.read()

content = content.replace("func TestStore_Load(t *testing.T) {\n\t\t\tt.Run(\"Happy Path\",", "func TestStore_Load(t *testing.T) {\n\tt.Run(\"Happy Path\",")

with open("server/pkg/storage/postgres/store_load_test.go", "w") as f:
    f.write(content)
