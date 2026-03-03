# Design Doc: Hardened Local Agent Sandbox
**Status:** Draft
**Created:** 2026-03-03

## 1. Context and Scope
As agentic tools like OpenClaw and Claude Code become ubiquitous, they introduce severe security risks when interacting with untrusted repositories. Attackers can embed malicious MCP configurations or lifecycle hooks that execute arbitrary code (RCE) on the developer's machine upon initialization.

MCP Any must evolve from a simple protocol proxy to a secure execution perimeter. The "Hardened Local Agent Sandbox" provides an isolated environment where local tools can be executed without risking the host system's integrity or sensitive credentials.

## 2. Goals & Non-Goals
* **Goals:**
    * Isolate local tool execution (shell, scripts, binaries) from the host filesystem and network.
    * Provide ephemeral, one-time-use execution environments for every tool call or session.
    * Support both Docker-based and WASM-based isolation layers.
    * Implement "Zero-Knowledge" credential passing, where tools only see specific, scoped environment variables.
* **Non-Goals:**
    * Providing a full GUI virtualization environment.
    * Replacing the user's primary shell for interactive manual use.

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Developer / "Vibe-Coder"
* **Primary Goal:** Safely explore and run tools from a newly cloned, untrusted GitHub repository.
* **The Happy Path (Tasks):**
    1. User points MCP Any to a new local project directory.
    2. MCP Any detects untrusted configuration hooks/tools.
    3. MCP Any prompts the user: "Run this tool in a Hardened Sandbox?"
    4. User approves. MCP Any spins up an ephemeral Docker container with a read-only mount of the project.
    5. The tool executes; its side effects are captured and presented to the user.
    6. The container is destroyed, leaving no persistent footprint on the host.

## 4. Design & Architecture
* **System Flow:**
    `LLM -> MCP Any Gateway -> Sandbox Manager -> Ephemeral Container (Tool) -> Result -> Gateway -> LLM`
* **APIs / Interfaces:**
    * `ISandboxProvider`: Interface for Docker, WASM, or Firecracker backends.
    * `SandboxConfig`: Defines resource limits (CPU/RAM), allowed network CIDRs, and read-only vs read-write mounts.
* **Data Storage/State:**
    * State is strictly ephemeral. Any persistent state must be explicitly committed back to the host via a "Verified Sync" hook.

## 5. Alternatives Considered
* **Host-Level Scoping (chroot/jails):** Rejected due to complexity in cross-platform (Windows/macOS) implementation and easier escape patterns compared to modern containers.
* **Pure WASM Execution:** Highly preferred for speed, but rejected as the *only* solution because many existing MCP tools rely on Python/Node/Shell environments not yet fully portable to WASM.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The sandbox follows the "Deny-All" network policy by default. DNS is restricted to verified upstream MCP servers only.
* **Observability:** Sandbox logs are streamed back to the MCP Any Audit Log with a `[SANDBOX]` prefix for clear attribution.

## 7. Evolutionary Changelog
* **2026-03-03:** Initial Document Creation.
