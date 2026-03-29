# Security Hardening: Path Traversal, Argument Injection, and Resource Limitations

## Vulnerability
1. **Prompt Injection / Execution Boundary Escape**: Found vulnerabilities in `validateSafePathAndInjection` inside `server/pkg/tool/types.go` where `strings.TrimSpace` was not being correctly applied to check logic like `checkForPathTraversal`, `checkForArgumentInjection`, and `checkForLocalFileAccess`. The presence of whitespace in payloads like ` " ../etc/passwd"` bypassed prefix checks, opening up Path Traversal and Command/Argument Injection vulnerabilities when resolving inputs for executor parameters.
2. **Resource Exhaustion**: Docker tool-executors previously ran without strict limits on memory or CPU, leading to potential denial of service if commands consume excessive resources.

## Exploitation Scenario
An attacker submitting tool inputs matching: `{"env_var": " -rm -rf"}` or `{"env_var": " ../etc/passwd"}` would successfully evade `validateSafePathAndInjection` checks due to lack of strict whitespace trimming before evaluating `strings.HasPrefix` logic. They could then achieve Argument Injection or Path Traversal into executor parameters.

## Remediation
- **Sandbox Hardening**: Applied `val = strings.TrimSpace(val)` in validation pipelines for `checkForPathTraversal`, `checkForLocalFileAccess`, and `checkForArgumentInjection` in `server/pkg/tool/types.go` to close whitespace evasion vectors.
- **Resource Limitations**: For Docker executors in `server/pkg/command/command.go`, added `Resources` block to `HostConfig` limiting container environments to `256MB` Memory and 1 CPU core `1000000000 NanoCPUs`, alongside a limit of 100 PIDs to harden against fork bombs or resource exhaustion scenarios.
- **Local Executor Safety**: Enforced strict `syscall.SysProcAttr{Setpgid: true}` limits on non-Windows targets for local commands via `exec.CommandContext` to prevent rogue child processes escaping execution sandboxes.
