import sys

with open('MODULE.bazel', 'r') as f:
    content = f.read()

content = content.replace("""go_deps.gazelle_override(
    path = "google.golang.org/genproto/googleapis/api",
)""", """go_deps.gazelle_override(
    directives = [
        "gazelle:resolve go google.golang.org/genproto/googleapis/api/annotations @org_golang_google_genproto_googleapis_api//annotations",
    ],
    path = "google.golang.org/genproto/googleapis/api",
)""")
with open('MODULE.bazel', 'w') as f:
    f.write(content)
