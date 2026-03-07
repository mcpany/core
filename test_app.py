import re

with open("server/pkg/app/server_test.go", "r") as f:
    content = f.read()

new_content = content.replace("app := NewApplication()", "app := NewApplication()\n\tapp.testMode = true")
with open("server/pkg/app/server_test.go", "w") as f:
    f.write(new_content)
