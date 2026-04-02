with open("server/pkg/app/server_init.go", "r") as f:
    content = f.read()
if "func (a *Application) seedTraces(" in content:
    print("seedTraces found")
else:
    print("seedTraces not found")
