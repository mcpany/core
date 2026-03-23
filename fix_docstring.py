import re

with open("server/pkg/storage/postgres/store_load_test.go", "r") as f:
    content = f.read()

content = content.replace("Summary: Validates that the store correctly loads and parses all server configuration from the database.", "Validates that the store correctly loads and parses all server configuration from the database.")
content = content.replace("TestStore_Load tests the Load method of the PostgreSQL store.\n//\n// Validates that the store correctly loads and parses all server configuration from the database.", "TestStore_Load tests the Load method of the PostgreSQL store.")

with open("server/pkg/storage/postgres/store_load_test.go", "w") as f:
    f.write(content)
