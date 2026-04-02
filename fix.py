with open("server/pkg/app/server_init.go", "r") as f:
    content = f.read()

content = content.replace("\t\tif err := a.seedTraces(ctx); err != nil {\n\t\tlog.Error(\"Failed to seed traces\", \"error\", err)\n\t}", "\tif err := a.seedTraces(ctx); err != nil {\n\t\tlog.Error(\"Failed to seed traces\", \"error\", err)\n\t}")

with open("server/pkg/app/server_init.go", "w") as f:
    f.write(content)
