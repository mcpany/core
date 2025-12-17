## 2025-02-18 - [Critical] Unrestricted Environment Variable Injection
**Vulnerability:** The `LocalCommandTool` and `CommandTool` implementations automatically injected all user inputs as environment variables into the executed command. This allowed attackers to inject dangerous variables like `LD_PRELOAD`, `PATH`, or application-specific overrides by simply providing them as tool inputs.
**Learning:** Implicitly mapping user inputs to the execution environment is unsafe. Convenience (passing all inputs as env vars) often leads to security vulnerabilities.
**Prevention:** Whitelist environment variables. Only pass inputs to the environment if they are explicitly defined in the configuration schema as parameters intended for the tool.

## 2025-12-17 - [Critical] Command Injection in Shell Wrapper
**Vulnerability:** The `buildCommandFromStdioConfig` function constructed a shell script by joining the command and arguments with spaces and passing it to `/bin/sh -c`. This allowed command injection if any argument contained shell metacharacters (e.g., `; rm -rf /`).
**Learning:** When using `sh -c` to chain commands (e.g., setup && run), simply joining arguments with spaces is unsafe.
**Prevention:** Always use `shellescape.Quote` (or equivalent) when interpolating arguments into a shell string.
