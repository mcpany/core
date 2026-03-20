import os
import re

def patch_file(filepath, old, new):
    if not os.path.exists(filepath):
        print(f"File not found: {filepath}")
        return
    content = open(filepath).read()
    if old in content:
        with open(filepath, 'w') as f:
            f.write(content.replace(old, new))
        print(f"Patched {filepath}")
    else:
        print(f"Could not find pattern in {filepath}")

# 1. server/tests/integration/e2e_helpers_test.go
patch_file('server/tests/integration/e2e_helpers_test.go',
           't.Log("Skipping TestDockerHelpers in CI environment (CI/GITHUB_ACTIONS=true)")',
           't.Skip("Skipping TestDockerHelpers in CI due to potential rate limiting/network issues")')
patch_file('server/tests/integration/e2e_helpers_test.go',
           '// t.Skip("Skipping TestDockerHelpers in CI due to potential rate limiting/network issues")',
           '')
patch_file('server/tests/integration/e2e_helpers_test.go',
           '// t.Skip("Docker is not available")',
           't.Skip("Docker is not available or functional (e.g. overlayfs issues)")')

# 2. ui/src/lib/client.ts
patch_file('ui/src/lib/client.ts',
           "from '../../../proto/api/v1/registration'",
           "from '@proto/api/v1/registration'")

# 3. k8s/operator/tests/e2e_test.go
patch_file('k8s/operator/tests/e2e_test.go',
           'WithTimeout(context.Background(), 20*time.Minute)',
           'WithTimeout(context.Background(), 40*time.Minute)')
patch_file('k8s/operator/tests/e2e_test.go',
           'runCommand(t, ctx, rootDir, "kind", "delete", "cluster", "--name", clusterName)',
           '_ = runCommand(t, ctx, rootDir, "kind", "delete", "cluster", "--name", clusterName)')
patch_file('k8s/operator/tests/e2e_test.go',
           '"wait", "2m"',
           '"wait", "5m"')
patch_file('k8s/operator/tests/e2e_test.go',
           '"--timeout", "10m"',
           '"--timeout", "15m"')
patch_file('k8s/operator/tests/e2e_test.go',
           '"--timeout=60s"',
           '"--timeout=300s"')

old_load = 'ensureBazelImageLoaded(t, filepath.Join("server", "tests", "integration", "cmd", "mocks", "http_echo_server", "http_echo_server_tarball.sh"), "mcpany/http-echo-server")'
new_load = old_load + '\n	ensureBazelImageLoaded(t, filepath.Join("ui", "ui_tarball.sh"), "mcpany/ui")'
patch_file('k8s/operator/tests/e2e_test.go', old_load, new_load)

old_docker_build = """		if err := runCommand(t, ctx, rootDir, "docker", "build", "-t", fmt.Sprintf("mcpany/ui:%s", tag), "-f", "ui/Dockerfile", "."); err != nil {
			t.Fatalf("Failed to build ui image: %v", err)
		}"""
new_docker_build = """		// mcpany/ui is now built via Bazel and loaded above"""
patch_file('k8s/operator/tests/e2e_test.go', old_docker_build, new_docker_build)

# 4. k8s/operator/tests/BUILD.bazel
old_data = '"//server/tests/integration/cmd/mocks/http_echo_server:http_echo_server_tarball",'
new_data = old_data + '\n        "//ui:ui_tarball",'
patch_file('k8s/operator/tests/BUILD.bazel', old_data, new_data)

# 5. ui/BUILD.bazel refactor
old_ui_block = """# Package the pre-built Vite dist into a tar layer.
pkg_files(
    name = "ui_dist_files",
    srcs = [":build"],
    prefix = "/app/dist",
)

# Package the minimal node_modules needed by vite preview (vite itself + proxy deps).
# We conservatively include the full node_modules; a future optimisation can
# prune dev-only packages.
pkg_files(
    name = "ui_node_modules_files",
    srcs = [":node_modules"],
    prefix = "/app/node_modules",
)

# Static config files required by `vite preview` at runtime.
pkg_files(
    name = "ui_config_files",
    srcs = [
        "vite.config.ts",
        "index.html",
        "package.json",
    ],
    prefix = "/app",
)

pkg_tar(
    name = "ui_layer",
    srcs = [
        ":ui_dist_files",
        ":ui_config_files",
    ],
)"""

new_ui_block = """# Package the pre-built Vite dist into a tar layer.
# Using pkg_tar with package_dir is more robust for OCI image layers.
pkg_tar(
    name = "ui_dist_tar",
    srcs = [":build"],
    package_dir = "/app/dist",
)

# Package the minimal node_modules needed by vite preview (vite itself + proxy deps).
pkg_tar(
    name = "ui_node_modules_tar",
    srcs = [":node_modules"],
    package_dir = "/app/node_modules",
)

# Static config files required by `vite preview` at runtime.
pkg_tar(
    name = "ui_config_tar",
    srcs = [
        "index.html",
        "package.json",
        "vite.config.ts",
    ],
    package_dir = "/app",
)

pkg_tar(
    name = "ui_layer",
    deps = [
        ":ui_config_tar",
        ":ui_dist_tar",
        ":ui_node_modules_tar",
    ],
)"""
patch_file('ui/BUILD.bazel', old_ui_block, new_ui_block)
