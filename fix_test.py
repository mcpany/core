import re

with open("server/pkg/storage/postgres/store_load_test.go", "r") as f:
    content = f.read()

# Make sure we don't redefine `err` if it's already defined from `sqlmock.New(...)` earlier in the function
content = content.replace("svcBytes, err := opts.Marshal(svc)", "svcBytes, err = opts.Marshal(svc)")
content = content.replace("userBytes, err := opts.Marshal(user)", "userBytes, err = opts.Marshal(user)")
content = content.replace("settingsBytes, err := opts.Marshal(settings)", "settingsBytes, err = opts.Marshal(settings)")
content = content.replace("collBytes, err := opts.Marshal(coll)", "collBytes, err = opts.Marshal(coll)")
content = content.replace("profileBytes, err := opts.Marshal(profile)", "profileBytes, err = opts.Marshal(profile)")

with open("server/pkg/storage/postgres/store_load_test.go", "w") as f:
    f.write(content)
