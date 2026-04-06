# Design Doc: Native Sandbox Adapter (NSA)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
AI agents are increasingly used to execute complex, multi-step tasks involving local tool execution (e.g., shell commands, code compilation, filesystem manipulation). Traditional process-level isolation is proving insufficient against sophisticated "Shadow-Sandbox" escapes. Gemini CLI v0.36.0 has set a new ecosystem baseline by integrating native OS sandboxing.

The Native Sandbox Adapter (NSA) will provide MCP Any with a unified, cross-platform interface for kernel-level tool isolation. By leveraging native OS primitives, we ensure that tools spawned by agents are restricted at the kernel level, independent of the agent framework's own limitations.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Implement a unified abstraction for macOS Seatbelt, Windows AppContainer, and Linux namespaces/cgroups.
    *   Provide declarative "Sandbox Profiles" for common tool categories (e.g., `web-browser`, `code-compiler`, `filesystem-analyzer`).
    *   Enforce absolute network and filesystem boundaries that survive subagent process spawns.
    *   Hardware-attest the active sandbox configuration before tool discovery.
*   **Non-Goals:**
    *   Replacing virtualization solutions like Docker for high-density multi-tenant environments (NSA is for local execution).
    *   Providing GUI sandboxing (focus is on CLI and background tools).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Local Agent Orchestrator
*   **Primary Goal:** Execute an untrusted "Data Conversion" script found in a GitHub repository without risking host-level file exfiltration.
*   **The Happy Path (Tasks):**
    1.  The Orchestrator identifies a tool call requiring local execution.
    2.  MCP Any selects the `restrictive-fs-only` NSA profile.
    3.  NSA initializes the kernel-level sandbox (e.g., Seatbelt on macOS).
    4.  The script executes. Any attempt to access `/Users/shared/secrets` or reach `evil.com` is denied by the OS kernel.
    5.  Tool output is returned, and the ephemeral sandbox is destroyed.

## 4. Design & Architecture
*   **System Flow:**
    `Orchestrator -> MCP Any (NSA) -> OS Kernel (Seatbelt/AppContainer) -> Tool Process`
*   **APIs / Interfaces:**
    *   `NSA.Initialize(profile_name)`
    *   `NSA.Execute(command, args)`
*   **Data Storage/State:**
    *   Sandbox profiles are stored in `server/config/sandbox/`.
    *   Execution logs include kernel-level denial events.

## 5. Alternatives Considered
*   **Docker-Only:** Rejected due to performance overhead and the "Double-Docker" complexity of running agents inside containers.
*   **gVisor:** Considered for Linux but lacks first-class macOS/Windows support in a lightweight local gateway context.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** Sandbox profiles must be TPM-signed to prevent "Profile Injection" attacks.
*   **Observability:** Kernel denial signals (e.g., `sandboxd` logs on macOS) must be surfaced to the MCP Any security dashboard.

## 7. Evolutionary Changelog
*   **2026-07-25:** Initial Document Creation.
