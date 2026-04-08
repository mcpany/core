import random
import os

docs = []
for root, dirs, files in os.walk('ui/docs'):
    for f in files:
        if f.endswith('.md'):
            docs.append(os.path.join(root, f))
for root, dirs, files in os.walk('server/docs'):
    for f in files:
        if f.endswith('.md'):
            docs.append(os.path.join(root, f))

# I will manually pick 10 varied files to ensure diversity.
files_to_check = [
    "ui/docs/features/webhooks.md",
    "ui/docs/features/test_connection.md",
    "server/docs/features/webhooks/README.md",
    "server/docs/features/health-checks.md",
    "server/docs/features/dynamic-ui.md",
    "server/docs/features/observability_guide.md",
    "server/docs/features/granular_scopes.md",
    "server/docs/features/shared_kv_store.md",
    "server/docs/features/authentication/README.md",
    "ui/docs/features/services.md"
]
for f in files_to_check:
    print(f"--- {f} ---")
    try:
        with open(f, 'r') as fp:
            print(fp.read()[:500])
    except:
        print("Not found")
