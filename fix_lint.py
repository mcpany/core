import json
with open("ui/package.json", "r") as f:
    pkg = json.load(f)

# eslint --quiet exit code might be non-zero if there are errors. We might have unaddressed errors in ui/src
# Let's see what pnpm lint really outputs
