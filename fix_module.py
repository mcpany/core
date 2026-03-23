import sys

filename = 'MODULE.bazel'
with open(filename, 'r') as f:
    lines = f.readlines()

start_index = -1
end_index = -1
for i, line in enumerate(lines):
    if 'use_repo(' in line and 'go_deps,' in line:
        start_index = i
    if start_index != -1 and ')' in line:
        end_index = i
        break

if start_index != -1 and end_index != -1:
    deps = set()
    for line in lines[start_index+1:end_index]:
        dep = line.strip().strip(',').strip('"')
        if dep:
            deps.add(dep)

    to_remove = ["com_github_mcpany_core", "io_k8s_api", "io_k8s_apimachinery", "io_k8s_sigs_controller_runtime"]
    for r in to_remove:
        if r in deps:
            deps.remove(r)

    deps.add("com_github_mcpany_core_proto")
    deps.add("org_golang_google_genproto_googleapis_api")

    sorted_deps = sorted(list(deps))
    new_lines = lines[:start_index+1]
    for d in sorted_deps:
        new_lines.append(f'    "{d}",\n')
    new_lines.extend(lines[end_index:])

    with open(filename, 'w') as f:
        f.writelines(new_lines)
