# Design Doc: Universal Skill Sandboxing
**Status:** Draft
**Created:** 2026-03-09

## 1. Context and Scope
With the explosion of the OpenClaw ecosystem and its "ClawHub" skill registry, thousands of unverified, third-party skills are being executed by local agents. Currently, these skills often inherit the full system permissions of the host agent, creating a massive security vulnerability (as seen in recent supply chain exploits). MCP Any must provide a "Zero-Trust Interception Layer" that sandboxes these skills, ensuring they can only access tools and resources explicitly granted by the user.

## 2. Goals & Non-Goals
* **Goals:**
    * Intercept all tool calls from agent-launched skills (specifically OpenClaw-compatible).
    * Enforce granular, per-skill permission scopes.
    * Provide a unified "Approval UI" for sensitive skill actions.
    * Maintain high performance with minimal latency overhead for local tool execution.
* **Non-Goals:**
    * Re-implementing a full virtual machine or container runtime (we leverage existing isolation where possible).
    * Policing the *content* of the skill (e.g., prompt injection) - we focus on the *capabilities* (tool access).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local AI Developer & Automation Power-User.
* **Primary Goal:** Safely run a new "GitHub Automation" skill from ClawHub without giving it access to local shell or private documents.
* **The Happy Path (Tasks):**
    1. User installs the skill via OpenClaw.
    2. OpenClaw routes skill tool-calls through MCP Any's Universal Sandbox.
    3. MCP Any detects a new skill identity and prompts the user for a "Capability Scope" (e.g., "Read-only access to GitHub tools").
    4. The skill attempts to call `local_shell.execute()`.
    5. MCP Any blocks the call as it's outside the granted scope and alerts the user.
    6. User approves the GitHub `read_repo` call, which proceeds.

## 4. Design & Architecture
* **System Flow:**
    `[OpenClaw Skill] -> [MCP Any Sandbox Proxy] -> [Policy Engine] -> [Verified MCP Server]`
* **APIs / Interfaces:**
    * `POST /v1/sandbox/execute`: Endpoint for routing skill tool-calls.
    * `GET /v1/sandbox/policies`: Query current skill-specific scopes.
* **Data Storage/State:**
    * SQLite-backed policy store mapping `skill_id` to `allowed_tool_regex`.

## 5. Alternatives Considered
* **OS-Level Sandboxing (Docker/Firejail):** Rejected for being too heavy and having high friction for casual users.
* **Token-Based Auth only:** Rejected because it doesn't solve the "unverified code" problem; a malicious skill with a valid token can still abuse its permissions.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All skills are "Untrusted" by default. Permission escalation requires explicit user interaction.
* **Observability:** Audit logs will record every tool call, its parameters, and the policy decision (Allow/Deny).

## 7. Evolutionary Changelog
* **2026-03-09:** Initial Document Creation.
