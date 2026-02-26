# Design Doc: Isolated Gateway Middleware (Sandbox)

**Status:** Draft
**Created:** 2026-02-26

## 1. Context and Scope
The rise of local-first agents (e.g., OpenClaw) has exposed a critical vulnerability: agents running with excessive local privileges. The "Lethal Trifecta" of LFI, prompt injection, and unauthorized host access allows malicious inputs to compromise the user's machine. MCP Any needs an "Isolated Gateway" that sandboxes tool execution to ensure that even if an agent is compromised, the host remains secure.

## 2. Goals & Non-Goals
* **Goals:**
    * Enforce process-level isolation for all command-line and stdio-based MCP tools.
    * Restrict filesystem access for tools to a pre-defined "Safe Zone" (e.g., `/app/sandbox`).
    * Implement resource limits (CPU, Memory) to prevent Denial of Service by rogue tools.
    * Provide "Ephemeral Execution" where each tool call runs in a clean, throwaway environment.
* **Non-Goals:**
    * Sandboxing remote HTTP-based MCP servers (these are isolated by network boundaries).
    * Replacing full-system virtualization (Docker/VMs) for all use cases (this is a lightweight middleware layer).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-conscious Developer using OpenClaw.
* **Primary Goal:** Execute a local script-analysis tool without risking access to private SSH keys.
* **The Happy Path (Tasks):**
    1. Developer enables `Isolated Gateway` in MCP Any config.
    2. OpenClaw requests execution of `analyze_script --file /home/user/project/script.py`.
    3. MCP Any spawns the `analyze_script` tool in a restricted namespace (e.g., using `chroot`, `unshare`, or a WASM runtime).
    4. The tool attempts to read `/home/user/.ssh/id_rsa` via an LFI exploit.
    5. The isolated environment blocks the access; MCP Any logs a security violation and returns an error to the agent.
    6. The host system remains untouched.

## 4. Design & Architecture
* **System Flow:**
    `Agent -> MCP Any (Policy Engine) -> Isolation Middleware -> [Isolated Subprocess] -> Tool Output -> MCP Any -> Agent`
* **APIs / Interfaces:**
    * Configuration: `isolation: { enabled: true, root: "/tmp/mcp-sandbox", max_memory_mb: 512 }`.
* **Data Storage/State:**
    * Ephemeral mounts are managed per-session.
    * Violation logs are stored in the Audit Log.

## 5. Alternatives Considered
* **Docker-per-call**: Too slow for high-frequency tool calls (latency > 1s).
* **WASM-only**: Many existing MCP tools are Python/Node based and cannot be easily ported to WASM.
* **NSJail/Bubblewrap**: Excellent for Linux, but requires a cross-platform abstraction for Windows (Job Objects) and macOS (Sandbox.kext).

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: The isolation layer is the "Last Line of Defense" if the Policy Firewall is bypassed via prompt injection.
* **Observability**: Real-time monitoring of sandboxed process health and resource usage in the UI.

## 7. Evolutionary Changelog
* **2026-02-26**: Initial Document Creation.
