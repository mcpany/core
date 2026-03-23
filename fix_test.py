import re

with open("server/pkg/storage/postgres/store_load_test.go", "r") as f:
    content = f.read()

# add period to docstring for side effects
content = content.replace("Side Effects:\n//   - Modifies testing state through assertions", "Side Effects:\n//   - Modifies testing state through assertions.")
with open("server/pkg/storage/postgres/store_load_test.go", "w") as f:
    f.write(content)
