# Design Doc: Discovery Sandbox Middleware

**Status:** Draft
**Created:** 2026-05-10

## 1. Context and Scope
The emergence of "Ghost-Execution" exploits in the Gemini CLI ecosystem has highlighted a critical vulnerability in how agent frameworks handle tool discovery. Configuration files (e.g., `.gemini/settings.json`) can define a `discoveryCommand` that executes automatically during the discovery phase. If a developer clones a malicious repository, these commands can execute with the developer's privileges before any explicit tool call is authorized. The Discovery Sandbox Middleware aims to isolate all discovery-time execution into a secure, ephemeral environment.

## 2. Goals & Non-Goals
* **Goals:**
    * Intercept all `discoveryCommand` or similar discovery-phase execution requests.
    * Execute discovery logic in a zero-trust, ephemeral sandbox (e.g., gVisor or isolated Docker container).
    * Prevent discovery-phase commands from accessing the host filesystem, network, or environment variables unless explicitly whitelisted.
    * Provide a mandatory attestation signal for discovered tools before they are exposed to the LLM.
* **Non-Goals:**
    * Sandboxing the primary tool execution (this is handled by the Policy Firewall and Detached Sandbox).
    * Modifying the behavior of the discovery commands themselves.

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Developer
* **Primary Goal:** Discover tools in a new repository without risking host compromise from malicious discovery hooks.
* **The Happy Path (Tasks):**
    1. The user opens a repository containing an MCP configuration with a `discoveryCommand`.
    2. MCP Any identifies the discovery request and routes it to the Discovery Sandbox Middleware.
    3. The Middleware spawns a "Ghost Shell" (isolated container) and executes the command.
    4. Discovered tool schemas are returned to MCP Any; the container is immediately destroyed.
    5. The user is notified that discovery was sandboxed and can review the resulting toolset.

## 4. Design & Architecture
* **System Flow:** The Discovery Manager delegates all external command execution to the Discovery Sandbox Middleware. The Middleware utilizes a lightweight virtualization layer (e.g., `bubblewrap` or `runnc`) to isolate the process from the host.
* **APIs / Interfaces:**
    * `ExecuteDiscovery(command, context)`: Entry point for discovery-phase execution.
    * `ValidateDiscoveryManifest(manifest)`: Post-discovery check to ensure the resulting tool schemas are structurally sound and free from "Context Poisoning."
* **Data Storage/State:** Ephemeral state only. No persistence allowed between discovery sessions.

## 5. Alternatives Considered
* **User Confirmation for Discovery**: Rejected due to "Approval Fatigue." Discovery happens frequently and users are likely to click through prompts.
* **Static Analysis of Discovery Commands**: Rejected because commands are often complex shell scripts that are difficult to analyze statically.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: The sandbox must be "Deny-by-Default" for all host resources.
* **Observability**: All discovery-phase execution logs are captured and available in the "Config Sandbox Monitor" UI.

## 7. Evolutionary Changelog
* **2026-05-10:** Initial Document Creation.
