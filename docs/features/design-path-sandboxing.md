# Design Doc: Non-Bypassable Path Sandboxing
**Status:** Draft
**Created:** 2026-03-08

## 1. Context and Scope
Recent vulnerabilities in the agent ecosystem (e.g., CVE-2026-25593 in OpenClaw) have demonstrated that AI agents with command-execution capabilities are highly susceptible to command injection via malicious configuration. If an attacker can manipulate the `cliPath` or equivalent configuration, they can achieve Remote Code Execution (RCE) with the privileges of the gateway process.

MCP Any, as a universal adapter, frequently uses the `Command` adapter to wrap local CLI tools. To prevent similar exploits, MCP Any must implement a hardened, non-bypassable path sandboxing mechanism that restricts execution to a verified set of binaries and paths.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Enforce a strict "Binary Whitelist" for all `Command` adapter instances.
    *   Implement a "Virtual Path Root" (Chroot-like behavior) for command execution.
    *   Prevent execution of any binary outside of `/usr/bin`, `/usr/local/bin`, or a user-defined `/opt/mcpany/bin`.
    *   Provide real-time rejection of configuration updates that attempt to use unverified paths.
*   **Non-Goals:**
    *   Providing full OS-level containerization (e.g., Docker) within the adapter itself (though it should complement such environments).
    *   Restricting arguments for whitelisted binaries (this is handled by the Policy Firewall).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Local LLM Swarm Orchestrator
*   **Primary Goal:** Safely wrap a local `git` and `ripgrep` tool for use by a subagent swarm.
*   **The Happy Path (Tasks):**
    1.  User adds a new `Command` service in `config.yaml` pointing to `/usr/bin/git`.
    2.  MCP Any Core validates `/usr/bin/git` against the internal whitelist and starts the service.
    3.  A malicious agent attempts to update the configuration to use `/tmp/malicious_bin`.
    4.  The Core's configuration validator rejects the change, logging a high-severity security event.

## 4. Design & Architecture
*   **System Flow:**
    - **Config Validation**: The `ServiceRegistry` uses a `PathValidator` middleware during service initialization and reload.
    - **Execution Guard**: Before `os/exec` is called, the `Command` adapter performs a final check against the resolved absolute path of the binary.
    - **Path Resolution**: Use `filepath.EvalSymlinks` to prevent symlink-based whitelist bypasses.
*   **APIs / Interfaces:**
    - `PathValidator` interface in `pkg/mcpserver`.
    - New config field: `security.allowed_binary_paths: []string` (with secure defaults).
*   **Data Storage/State:** The whitelist is loaded from the global configuration and immutable during runtime unless a signed configuration update is received.

## 5. Alternatives Considered
*   **Simple Regex Validation**: *Rejected* because regex is notoriously easy to bypass with relative paths (`../`) or symlinks.
*   **Manual Approval for every Tool**: *Rejected* as it adds too much friction for developers. Whitelisting provides a better balance of safety and speed.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** This is a core "Safe-by-Default" feature. It ensures that even if an agent compromises the configuration API, the blast radius is limited to pre-approved binaries.
*   **Observability:** Security violations must be logged with a unique error code (e.g., `ERR_PATH_VIOLATION`) and immediately visible in the UI's Security Dashboard.

## 7. Evolutionary Changelog
*   **2026-03-08:** Initial Document Creation (Response to CVE-2026-25593).
