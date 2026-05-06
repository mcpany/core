# Friction Audit & Impact Report: MCP Any Configuration Setup

## Friction Identified
During a "Day One" experience simulation to launch the "Hello World" example via the CLI, the following issues were encountered:
1. **CLI Flag Discrepancy**: The `README.md` documented using `-config examples/hello_world.yaml`, but the CLI (built with Cobra/Viper) strictly expects `--config`. This caused an `unknown command` error immediately upon copy-pasting the instructions.
2. **Relative Path Resolution Failure with Bazel**: Even with the correct flag (`--config`), running via `bazelisk run` executes the binary from within a Bazel sandbox/execroot. As a result, the relative path `examples/hello_world.yaml` failed to resolve, raising a `❌ Configuration file not found` error.

## Engineering Fixes Applied
1. **Dynamic Path Resolution**: Refactored the `getStringSlice` helper in `server/pkg/config/settings.go`. The logic was updated to intercept `config-path` and `config` keys and automatically resolve relative paths using the `BUILD_WORKSPACE_DIRECTORY` environment variable, which Bazel populates during `bazel run`.
2. **Robust Array Parsing**: Maintained and strengthened Viper's array-parsing coverage. The `getStringSlice` implementation was carefully tuned to correctly handle both string-based comma-separated lists and multi-element arrays with nested commas (which are expected by the `config_test` coverage tests).
3. **Documentation Alignment**: Corrected the `README.md` to properly use `--config` (and explicitly target the `run` command) so the "Hello World" snippet correctly executes.

## Result & Verification
* **Measurable Reduction in Friction**: Developers no longer have to manually construct absolute paths or guess the correct CLI subcommands. They can successfully execute the provided "Hello World" instruction directly from the project root without encountering `file not found` errors.
* **Test Verification**: All existing tests, including robust config generation and edge-case testing, continue to pass 100% green via `bazelisk test //...`. (Note: Resolved an existing race condition in `server/pkg/bus/redis/bus_test.go` using atomic counters to ensure a hermetic testing environment).
