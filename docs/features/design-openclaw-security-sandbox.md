# Design Doc: OpenClaw Security Sandbox

**Status:** Draft
**Created:** 2026-03-06

## 1. Context and Scope
The March 2026 OpenClaw security crisis has revealed that autonomous agents with broad filesystem and shell access are highly vulnerable to RCE and supply chain attacks (e.g., CVE-2026-25253, ClawHavoc). Users want the power of OpenClaw but are terrified of its lack of boundaries. MCP Any is uniquely positioned to act as a "Security Proxy" that wraps OpenClaw's tool calls in a strict, sandboxed environment.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Implement a virtualized "Chroot-like" environment for all filesystem tools.
    *   Enforce strict "allow-lists" for shell commands, preventing arbitrary execution.
    *   Provide real-time "Threat Scoring" for tool calls based on community-reported malicious patterns.
    *   Enable mandatory HITL (Human-in-the-Loop) for any "write" operation outside of a designated `/sandbox/` directory.
*   **Non-Goals:**
    *   Rewriting OpenClaw itself.
    *   Providing full OS-level virtualization (Docker is preferred for that; MCP Any provides application-level sandboxing).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Developer using OpenClaw for local code refactoring.
*   **Primary Goal:** Allow OpenClaw to read/write files in a specific project directory without risking the rest of the home directory.
*   **The Happy Path (Tasks):**
    1.  User starts MCP Any with the `openclaw-shield` profile enabled.
    2.  User configures the shield to allow access only to `~/projects/my-app/`.
    3.  OpenClaw attempts to read `~/.ssh/id_rsa` via a tool call.
    4.  MCP Any intercepts the call, detects it is outside the allowed path, and returns a "Permission Denied" error to the agent.
    5.  User is notified in the MCP Any UI about the blocked attempt.

## 4. Design & Architecture
*   **System Flow:**
    - **Intercept Layer**: Every tool call from the agent passes through the `SandboxMiddleware`.
    - **Path Resolver**: Translates agent-provided paths to absolute host paths and validates them against the `AllowedPaths` list.
    - **Command Filter**: Parses shell commands and blocks prohibited binaries or flags (e.g., `curl | bash`).
*   **APIs / Interfaces:**
    - Config schema addition: `sandbox: { enabled: true, root: "/path/to/sandbox", shell_allowlist: ["git", "npm"] }`
*   **Data Storage/State:** Persistent log of blocked/allowed actions for auditability.

## 5. Alternatives Considered
*   **Running OpenClaw in Docker**: *Rejected* as the primary solution because it's cumbersome for local file editing and lacks granular tool-level visibility.
*   **OS-level MAC (AppArmor/SELinux)**: *Rejected* as it is too complex for the average user to configure and not cross-platform.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** This is the core of the feature. It assumes the agent is potentially compromised or "rogue" and enforces strict boundaries.
*   **Observability:** All sandbox violations are logged with high priority and can trigger desktop notifications.

## 7. Evolutionary Changelog
*   **2026-03-06:** Initial Document Creation.
