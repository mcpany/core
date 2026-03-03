# Design Doc: Project-Config Sandbox Middleware

**Status:** Draft
**Created:** 2026-03-03

## 1. Context and Scope
With the rise of agentic tools like Claude Code and OpenClaw, it is increasingly common for agents to operate within repositories that contain their own tool configurations (e.g., `.mcp` files, project-level hooks). Recent vulnerabilities (CVE-2025-59536) have shown that these project-controlled configurations are a massive attack vector for Remote Code Execution (RCE) and credential exfiltration.

MCP Any needs a middleware layer that "sandboxes" these project-level configurations, ensuring they cannot execute arbitrary code or access sensitive host resources without explicit, validated permission.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Intercept all project-level tool configuration loading attempts.
    *   Validate configurations against a strict, non-executable JSON/YAML schema.
    *   Isolate project-defined tools into a "Project-Scoped" execution context.
    *   Prevent project-level configs from overriding global security policies or environment variables.
    *   Require explicit user "Opt-In" before activating tools from an untrusted repository.
*   **Non-Goals:**
    *   Replacing global configuration (this middleware only handles project-specific overrides).
    *   Providing a full OS-level sandbox (focus is on config-level sandboxing).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Developer cloning an open-source repository.
*   **Primary Goal:** Allow the agent to use project-specific tools safely without risking machine takeover.
*   **The Happy Path (Tasks):**
    1.  User clones a repo and runs an agent with MCP Any.
    2.  MCP Any detects a project-level configuration (e.g., `mcp.json`).
    3.  The **Project-Config Sandbox** intercepts the load and flags "New untrusted configuration found."
    4.  The user is presented with a simplified, safe view of what the config wants to do (e.g., "Add tool: 'lint-helper'").
    5.  User approves the specific tools.
    6.  MCP Any loads the tools in a restricted mode where `env` access and `shell` execution are strictly governed by the global Policy Firewall.

## 4. Design & Architecture
*   **System Flow:**
    - **Interception Layer**: The `ConfigLoader` is wrapped by the `SandboxMiddleware`.
    - **Validation Engine**: Uses a JSON Schema validator to ensure no hidden `hooks` or `pre-exec` scripts are present in the config.
    - **Context Isolation**: Tools from project configs are tagged with `origin: project` and restricted to the project root directory.
*   **APIs / Interfaces:**
    - `POST /v1/config/project/validate`: Validates a proposed project config.
    - `POST /v1/config/project/approve`: Mark a project config as trusted for the current session.
*   **Data Storage/State:** A "Trust Registry" stores hashes of approved project configurations to prevent re-approval fatigue.

## 5. Alternatives Considered
*   **Complete Disabling of Project Configs**: Too restrictive for developer productivity.
*   **Manual Code Review of Configs**: Unrealistic for agents operating autonomously; the system must provide the first line of defense.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** This is a direct response to the "Configuration Injection" threat model. It implements "Least Privilege" for project-level assets.
*   **Observability:** All blocked configuration attempts must be logged with specific reasons (e.g., "Blocked: found 'shell' hook in untrusted config").

## 7. Evolutionary Changelog
*   **2026-03-03:** Initial Document Creation.
