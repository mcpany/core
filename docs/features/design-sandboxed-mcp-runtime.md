# Design Doc: Sandboxed MCP Runtime

**Status:** Draft
**Created:** 2026-02-27

## 1. Context and Scope
Recent security vulnerabilities in AI coding assistants (e.g., Claude Code CVE-2026-21852) have demonstrated that executing MCP servers and lifecycle hooks directly on a developer's host machine is inherently risky. A malicious repository can configure an MCP server to execute arbitrary shell commands or exfiltrate sensitive data (API keys, env vars) during initialization. MCP Any must provide a secure, isolated runtime for these tools to mitigate host-level exploitation.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Isolate MCP server execution from the host filesystem and network.
    *   Support Docker-based and WASM-based sandboxing.
    *   Provide a "Bridge" for controlled resource access (e.g., specific directory mounts).
    *   Implement "Admin Allowlisting" for host-access tools.
*   **Non-Goals:**
    *   Providing a full virtual desktop environment.
    *   Securing the LLM itself (focus is on tool execution).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Security-Conscious Developer.
*   **Primary Goal:** Run a tool from an untrusted GitHub repository without risking host machine takeover.
*   **The Happy Path (Tasks):**
    1.  User points MCP Any to a repository containing an `mcp.yaml`.
    2.  MCP Any detects the tools and hooks.
    3.  Because the repo is "untrusted," MCP Any initializes a lightweight Docker container for the tools.
    4.  The tools execute within the container, with only the repository folder mounted as read-only.
    5.  Any attempt to access `~/.ssh` or environment variables outside the container is blocked by the sandbox.

## 4. Design & Architecture
*   **System Flow:**
    - **Runtime Selector**: A middleware that determines the execution environment (Host vs. Sandbox) based on the "Trust Tier" of the configuration.
    - **Sandbox Provider**: Abstraction layer for `DockerRuntime` or `WasmRuntime`.
    - **Resource Mapping**: Explicit configuration of which host resources (files, env vars) are passed into the sandbox.
*   **APIs / Interfaces:**
    - `RuntimeProvider` interface with `Start()`, `Stop()`, and `Execute()` methods.
    - `SandboxConfig` schema for defining resource constraints (CPU, Memory, Network).
*   **Data Storage/State:** Sandbox definitions are persisted in the global MCP Any configuration, but runtime state is ephemeral.

## 5. Alternatives Considered
*   **Process-level Isolation (chroot/jails)**: Harder to manage cross-platform (Windows/macOS). *Rejected* in favor of Docker/WASM for better portability.
*   **User Confirmation for Every Call**: Too much friction for developers (HITL fatigue). *Rejected* in favor of "Secure by Default" sandboxing.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** The sandbox is the primary enforcement point. Even if an LLM is "tricked" into a malicious call, the sandbox prevents host damage.
*   **Observability:** Log all sandbox violations (e.g., blocked syscalls or network attempts) to the Security Dashboard.

## 7. Evolutionary Changelog
*   **2026-02-27:** Initial Document Creation.
