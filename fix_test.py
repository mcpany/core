import re

with open("server/pkg/storage/postgres/store_load_test.go", "r") as f:
    content = f.read()

content = content.replace("func TestStore_Load(t *testing.T) {\n", """// TestStore_Load tests the Load method of the PostgreSQL store.
//
// Parameters:
//   - t (*testing.T): The testing context.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - Modifies testing state through assertions.
func TestStore_Load(t *testing.T) {\n""")

with open("server/pkg/storage/postgres/store_load_test.go", "w") as f:
    f.write(content)
