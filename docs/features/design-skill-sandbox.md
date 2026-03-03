# Design Doc: Ephemeral Skill Sandbox (Tool Hypervisor)

**Status:** Draft
**Created:** 2026-03-03

## 1. Context and Scope
The OpenClaw RCE vulnerability (CVE-2026-25253) highlighted the risk of running AI agent "skills" (MCP servers) directly on the host machine. If a tool has a vulnerability, or if an agent is tricked via prompt injection into calling a tool with malicious arguments, the entire host is compromised. MCP Any must provide an isolation layer that treats every MCP server as an untrusted guest.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Provide multiple isolation backends: Subprocess (default), Docker/Container, and WASM.
    *   Enforce resource limits (CPU, Memory, Network) per tool execution.
    *   Implement "Ephemeral-by-Default": sandbox environments are destroyed after the tool call or session ends.
    *   Standardize filesystem access via virtualized mounts (restricted to specific directories).
*   **Non-Goals:**
    *   Solving all possible side-channel attacks.
    *   Replacing the host operating system's security model.

## 3. Critical User Journey (CUJ)
*   **User Persona:** Enterprise Security Admin.
*   **Primary Goal:** Allow developers to use the "PostgreSQL Tools" MCP server without giving the server process access to the host's `/etc/shadow` file.
*   **The Happy Path (Tasks):**
    1.  Admin configures the PostgreSQL MCP server to run in the `docker` sandbox.
    2.  MCP Any pulls the hardened tool image.
    3.  When an agent calls a tool, MCP Any starts a transient container with restricted networking and only the necessary database credentials injected.
    4.  The tool executes, returns the result, and the container is immediately SIGKILLed.

## 4. Design & Architecture
*   **System Flow:**
    - **Sandbox Manager**: Orchestrates the lifecycle of different sandbox backends.
    - **Isolation Layer**: Wraps the MCP transport (Stdio/HTTP) with a sandboxing wrapper.
    - **Policy Enforcement**: Integrates with the `Policy Firewall` to determine which sandbox profile to apply to a given tool.
*   **APIs / Interfaces:**
    - `SandboxProvider` interface in Go/TypeScript.
    - Config fields for `sandbox_profile` in `mcp.yaml`.
*   **Data Storage/State:** Ephemeral state only. Any persistent state must go through the `Shared KV Store`.

## 5. Alternatives Considered
*   **Manual OS-level Sandboxing (chroot/jails)**: Hard for users to configure. *Rejected* in favor of containerization.
*   **WASM-Only**: Highly secure but many existing MCP servers are written in Python/Node and don't yet support WASM. *Rejected* as the *only* option, but included as a high-security backend.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** This is the "Defense in Depth" layer. Even if the Policy Firewall is bypassed, the Sandbox prevents host-level damage.
*   **Observability:** Track sandbox "Violations" (e.g., attempt to access blocked files) in the security dashboard.

## 7. Evolutionary Changelog
*   **2026-03-03:** Initial Document Creation.
