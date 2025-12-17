## 2025-02-18 - [Critical] Unrestricted Environment Variable Injection
**Vulnerability:** The `LocalCommandTool` and `CommandTool` implementations automatically injected all user inputs as environment variables into the executed command. This allowed attackers to inject dangerous variables like `LD_PRELOAD`, `PATH`, or application-specific overrides by simply providing them as tool inputs.
**Learning:** Implicitly mapping user inputs to the execution environment is unsafe. Convenience (passing all inputs as env vars) often leads to security vulnerabilities.
**Prevention:** Whitelist environment variables. Only pass inputs to the environment if they are explicitly defined in the configuration schema as parameters intended for the tool.
