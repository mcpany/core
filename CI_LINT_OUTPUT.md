make -C server lint
make[1]: Entering directory '/app/server'
INFO: Invocation ID: 861e4fa7-45f9-4c3a-b93b-90148800598b
INFO: Options provided by the client:
  Inherited 'common' options: --isatty=0 --terminal_columns=80
INFO: Reading rc options for 'run' from /app/.bazelrc:
  Inherited 'build' options: --disk_cache=~/.cache/bazel-disk-cache --action_env=CGO_ENABLED=1 --spawn_strategy=sandboxed,local --incompatible_strict_action_env --announce_rc --action_env=GEMINI_API_KEY --action_env=MCP_BUNDLE_DIR --action_env=CI
Computing main repo mapping:
Loading:
Loading: 0 packages loaded
Analyzing: target //:lint (0 packages loaded, 0 targets configured)
Analyzing: target //:lint (0 packages loaded, 0 targets configured)

INFO: Analyzed target //:lint (0 packages loaded, 2 targets configured).
INFO: Found 1 target...
Target //:lint up-to-date:
  ../bazel-bin/lint
INFO: Elapsed time: 0.514s, Critical Path: 0.01s
INFO: 1 process: 5 action cache hit, 1 internal.
INFO: Build completed successfully, 1 total action
INFO: Running command line: ../bazel-bin/lint
Running Gazelle...
INFO: Invocation ID: 177c383a-c744-4b1c-9039-009b075a7752
INFO: Options provided by the client:
  Inherited 'common' options: --isatty=0 --terminal_columns=80
INFO: Reading rc options for 'run' from /app/.bazelrc:
  Inherited 'build' options: --disk_cache=~/.cache/bazel-disk-cache --action_env=CGO_ENABLED=1 --spawn_strategy=sandboxed,local --incompatible_strict_action_env --announce_rc --action_env=GEMINI_API_KEY --action_env=MCP_BUNDLE_DIR --action_env=CI
Computing main repo mapping:
Loading:
Loading: 0 packages loaded
Analyzing: target //:gazelle (0 packages loaded, 0 targets configured)
Analyzing: target //:gazelle (0 packages loaded, 0 targets configured)

Analyzing: target //:gazelle (51 packages loaded, 9526 targets configured)

INFO: Analyzed target //:gazelle (57 packages loaded, 11028 targets configured).
INFO: Found 1 target...
Target //:gazelle up-to-date:
  bazel-bin/gazelle-runner.bash
  bazel-bin/gazelle
INFO: Elapsed time: 2.251s, Critical Path: 0.09s
INFO: 1 process: 72 action cache hit, 1 internal.
INFO: Build completed successfully, 1 total action
INFO: Running command line: bazel-bin/gazelle
Running Buildifier...
INFO: Invocation ID: 71b099ce-5d3e-4571-9ecb-1891e1cd527a
INFO: Options provided by the client:
  Inherited 'common' options: --isatty=0 --terminal_columns=80
INFO: Reading rc options for 'run' from /app/.bazelrc:
  Inherited 'build' options: --disk_cache=~/.cache/bazel-disk-cache --action_env=CGO_ENABLED=1 --spawn_strategy=sandboxed,local --incompatible_strict_action_env --announce_rc --action_env=GEMINI_API_KEY --action_env=MCP_BUNDLE_DIR --action_env=CI
Computing main repo mapping:
Loading:
Loading: 0 packages loaded
Analyzing: target //:buildifier (0 packages loaded, 0 targets configured)
Analyzing: target //:buildifier (0 packages loaded, 0 targets configured)

INFO: Analyzed target //:buildifier (2 packages loaded, 7 targets configured).
INFO: Found 1 target...
Target @@buildifier_prebuilt+//buildifier:buildifier up-to-date:
  bazel-bin/external/buildifier_prebuilt+/buildifier/buildifier
INFO: Elapsed time: 0.534s, Critical Path: 0.00s
INFO: 1 process: 5 action cache hit, 1 internal.
INFO: Build completed successfully, 1 total action
INFO: Running command line: bazel-bin/external/buildifier_prebuilt+/buildifier/buildifier <args omitted>
Running golangci-lint...
Running pre-commit...
check for added large files..............................................Passed
check for case conflicts.................................................Passed
check json...............................................................Passed
check yaml...............................................................Passed
detect private key.......................................................Passed
fix end of files.........................................................Passed
trim trailing whitespace.................................................Passed
shellcheck...............................................................Passed
helm-lint................................................................Passed
check-go-doc.............................................................Passed
check-ts-doc.............................................................Passed
addlicense...............................................................Passed
make[1]: Leaving directory '/app/server'
