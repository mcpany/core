import re

with open("server/pkg/storage/postgres/store_load_test.go", "r") as f:
    content = f.read()

content = content.replace("// TestStore_Load tests the Load method of the PostgreSQL store.\n//\n// Parameters:\n//   - t (*testing.T): The testing context.\n//\n// Returns:\n//   - None.\n//\n// Errors:\n//   - None.\n//\n// Side Effects:\n//   - Modifies testing state through assertions.", "// TestStore_Load tests the Load method of the PostgreSQL store.\n//\n// Summary: Validates that the store correctly loads and parses all server configuration from the database.\n//\n// Parameters:\n//   - t (*testing.T): The testing context.\n//\n// Returns:\n//   - None.\n//\n// Errors:\n//   - None.\n//\n// Side Effects:\n//   - Modifies testing state through assertions.")
content = content.replace("func TestStore_Load(t *testing.T) {\n\t// Summary: Validates that the store correctly loads and parses all server configuration from the database.\n\tt.Run(\"Happy Path\",", "func TestStore_Load(t *testing.T) {\n\tt.Run(\"Happy Path\",")

with open("server/pkg/storage/postgres/store_load_test.go", "w") as f:
    f.write(content)
