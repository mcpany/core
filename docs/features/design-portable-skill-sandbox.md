# Design Doc: Portable Skill Sandbox & Signature Matching
**Status:** Draft
**Created:** 2026-03-11

## 1. Context and Scope
The "ToxicSkills" research has revealed that malicious AI agent skills (tools) are being designed to be portable across different agent frameworks (OpenClaw, Cursor, Claude Code, etc.). These skills often bypass traditional security checks by appearing innocuous while performing background data exfiltration or RCE. MCP Any needs a universal way to identify, sandbox, and block these portable threats regardless of which agent framework is being used.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a "Cross-Platform Skill Signature Matching" engine that identifies known malicious tool patterns.
    * Provide a mandatory "Detached Sandbox" for all non-attested tool executions.
    * Establish a "Global Attestation Mesh" to share trust data with other MCP Any instances.
    * Enforce strict Origin and CSRF validation for all WebSocket-based tool management.
* **Non-Goals:**
    * Creating a separate marketplace for skills (MCP Any remains a gateway/adapter).
    * Blocking all unverified tools (instead, sandbox them by default).

## 3. Critical User Journey (CUJ)
* **User Persona:** Developer using multiple AI agents (e.g., OpenClaw for personal tasks, Claude Code for work).
* **Primary Goal:** Ensure that a "ToxicSkill" discovered in the OpenClaw ecosystem is automatically blocked or sandboxed when encountered in Claude Code via MCP Any.
* **The Happy Path (Tasks):**
    1. User attempts to install a new MCP tool suggested by an agent.
    2. MCP Any intercepts the tool registration.
    3. The `Signature Matcher` identifies the tool's code pattern as matching a known "ToxicSkill" signature from the Global Attestation Mesh.
    4. MCP Any blocks the registration and alerts the user with a "High Risk" warning.
    5. If the user insists on using a new, unknown tool, MCP Any forces it to run in a `Detached Sandbox` with zero network access.

## 4. Design & Architecture
* **System Flow:**
    `Agent` -> `MCP Any Gateway` -> `Signature Matcher` -> `Sandbox Manager` -> `Tool Execution`
    1. **Pattern Matching**: Every tool definition (JSON schema + implementation) is hashed and checked against a local cache of "Toxic signatures."
    2. **Mesh Sync**: The local cache is periodically updated from the Global Attestation Mesh (federated MCP Any nodes).
    3. **Sandboxing**: The `Sandbox Manager` uses OCI-compliant containers (Docker/Podman) or gVisor to isolate tool execution.
* **APIs / Interfaces:**
    * `GET /v1/security/signatures`: Fetch latest toxic signatures.
    * `POST /v1/security/attest`: Submit a manual attestation for an unknown tool.
* **Data Storage/State:**
    * `signatures.db`: SQLite database of known malicious tool patterns and trust scores.

## 5. Alternatives Considered
* **Static Analysis Only**: Rejected because malicious payloads are often obfuscated or generated dynamically. Sandboxing is required for "Defense in Depth."
* **Centralized Registry**: Rejected in favor of a Federated Mesh to avoid a single point of failure and resist censorship.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: All tool executions are sandboxed by default. WebSockets use mandatory HMAC-based authentication and Origin headers.
* **Performance**: Signature matching is performed asynchronously during tool discovery to minimize latency. Sandboxing adds minimal overhead (~10ms) using pre-warmed containers.

## 7. Evolutionary Changelog
* **2026-03-11:** Initial Document Creation.
