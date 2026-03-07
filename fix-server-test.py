import re

with open("server/pkg/app/server_test.go", "r") as f:
    content = f.read()

# remove all app.testMode = true lines
new_content = re.sub(r'\s*app\.testMode = true\n', '\n', content)

with open("server/pkg/app/server_test.go", "w") as f:
    f.write(new_content)
