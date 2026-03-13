# Design Doc: Command-Adapter Behavioral Attestation
**Status:** Draft
**Created:** 2026-03-25

## 1. Context and Scope
The discovery of CVE-2026-0755 in the `gemini-mcp-tool` ecosystem has highlighted a critical flaw in how AI agents interact with local CLI tools. Static sanitization is proving insufficient against sophisticated command injection patterns. Command-Adapter Behavioral Attestation introduces a "Ghost Shell" profiling layer that executes and monitors command tools in an isolated sandbox before they are allowed to run on the host system.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement "Ghost Shell" behavioral profiling for all CLI-based tools.
    * Detect and block unauthorized file access, network connections, or shell escapes during the profiling phase.
    * Provide a cryptographic "Behavioral Attestation" token for approved commands.
* **Non-Goals:**
    * Provide a permanent sandbox for all tool execution (Ghost Shell is for profiling/attestation).
    * Replace existing filesystem permissions (UID/GID).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local Agent Power User
* **Primary Goal:** Use a third-party CLI tool via MCP without risking a command injection attack that could wipe the home directory.
* **The Happy Path (Tasks):**
    1. User configures a new `Command` adapter for a tool (e.g., `process-csv`).
    2. Agent attempts to call the tool with specific arguments.
    3. MCP Any intercepts the call and spawns a "Ghost Shell" (isolated container/sandbox).
    4. The command is executed in the Ghost Shell while syscalls and network activity are monitored.
    5. The profiling engine confirms the command behavior matches the tool's declared scope (e.g., only reads CSV files).
    6. MCP Any issues a "Behavioral Attestation" and allows the command to run on the host.

## 4. Design & Architecture
* **System Flow:**
    `[MCP Call] -> [Ghost Shell Sandbox] -> [Behavioral Profiler] -> [Attestation Engine] -> [Host Execution]`
* **APIs / Interfaces:**
    * Internal `ProfileCommand(cmd, args)` service.
    * `GhostShellInterface` for pluggable sandbox backends (Docker, gVisor, WASM).
* **Data Storage/State:**
    * Cache of attested command signatures (SHA-256 of cmd + args + behavior profile).

## 5. Alternatives Considered
* **Regex-based Sanitization:** Rejected as it failed to prevent CVE-2026-0755.
* **Manual Command Approval:** Rejected as it significantly degrades the user experience for high-frequency agents.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The Ghost Shell itself must be treated as untrusted and resource-constrained.
* **Observability:** Detailed logs of blocked syscalls or network attempts during profiling.

## 7. Evolutionary Changelog
* **2026-03-25:** Initial Document Creation.
