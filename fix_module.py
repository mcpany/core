import sys

filename = 'MODULE.bazel'
with open(filename, 'r') as f:
    content = f.read()

# Find use_repo(go_deps, ...)
import re
match = re.search(r'use_repo\(\s*go_deps,(.*?)\)', content, re.DOTALL)
if match:
    deps_str = match.group(1)
    deps = set()
    for d in re.findall(r'"(.*?)"', deps_str):
        deps.add(d)

    # Critical removals
    to_remove = ["com_github_mcpany_core", "io_k8s_api", "io_k8s_apimachinery", "io_k8s_sigs_controller_runtime"]
    for r in to_remove:
        if r in deps:
            deps.remove(r)

    # Required additions
    deps.add("com_github_mcpany_core_proto")
    deps.add("org_golang_google_genproto_googleapis_api")

    sorted_deps = sorted(list(deps))
    new_deps_str = '\n    ' + ',\n    '.join(f'"{d}"' for d in sorted_deps) + ',\n'

    new_content = content[:match.start(1)] + new_deps_str + content[match.end(1):]
    with open(filename, 'w') as f:
        f.write(new_content)
