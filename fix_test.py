import re

with open("server/pkg/storage/postgres/store_load_test.go", "r") as f:
    content = f.read()

content = content.replace("func TestStore_Load(t *testing.T) {\n\t// Use MatchExpectationsInOrder(false) because go routines can query in any order\n\t// Actually no, just simple mock with MatchAnyArgs will do, but here order is generic", "func TestStore_Load(t *testing.T) {")

with open("server/pkg/storage/postgres/store_load_test.go", "w") as f:
    f.write(content)
