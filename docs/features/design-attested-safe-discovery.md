# Design Doc: Project-Scoped Safe Discovery (Attested Safe Discovery)
**Status:** Draft
**Created:** 2026-03-01

## 1. Context and Scope
With the rise of agentic coding assistants like Claude Code and OpenClaw, there is a growing trend of "project-specific" MCP servers and configurations (e.g., defined in `.claude/settings.json` or `.mcp/config.yaml`). However, recent security vulnerabilities (CVE-2025-59536) have shown that automatically loading these configurations can lead to Remote Code Execution (RCE) and API key exfiltration if a user opens a malicious repository.

MCP Any needs to provide a secure, sandboxed bridge for project-scoped tools that ensures no tool is executed or discovered without explicit attestation and security boundary enforcement.

## 2. Goals & Non-Goals
* **Goals:**
    * Prevent automatic execution of unverified shell commands or MCP servers from local project files.
    * Enforce a "Zero Trust" boundary for project-scoped tools.
    * Provide a mechanism for "Hierarchical Attestation" where a user or parent agent must vouch for a project configuration.
    * Isolate environment variables to prevent exfiltration.
* **Non-Goals:**
    * Implementing a full sandbox for the tool itself (that is the responsibility of the Upstream Adapter/Runtime).
    * Replacing existing MCP server discovery mechanisms (this acts as a security middleware).

## 3. Critical User Journey (CUJ)
* **User Persona:** Developer using a multi-agent swarm (OpenClaw) on a new project.
* **Primary Goal:** Use project-specific tools (e.g., a custom linter MCP) safely without risking RCE from hidden malicious configs.
* **The Happy Path (Tasks):**
    1. User clones a repository containing a `.mcpany/project-tools.yaml` file.
    2. MCP Any detects the project context but marks all tools as `QUARANTINED`.
    3. The agent requests a tool from the project scope.
    4. MCP Any triggers a `HITL` (Human-In-The-Loop) or `Attestation` request.
    5. User verifies the tool signature or command-line template.
    6. MCP Any promotes the tool to `ATTESTED` and allows execution within a restricted sub-process.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        Repo[.mcpany/config.yaml] -->|Load| Scanner[Project Scanner]
        Scanner -->|Quarantine| Registry[Service Registry]
        Registry -->|Policy Check| Engine[Policy Engine]
        Engine -->|Request Attestation| UI[User/Parent Agent]
        UI -->|Signed Token| Engine
        Engine -->|Promote| Registry
        Registry -->|Execute| Adapter[Command/HTTP Adapter]
    ```
* **APIs / Interfaces:**
    * `DiscoveryService.RegisterProject(path string) (ProjectID, error)`
    * `PolicyEngine.AttestTool(toolID string, signature string) error`
* **Data Storage/State:**
    * Project configurations are stored in the Service Registry with a `TrustLevel` metadata field (`UNTRUSTED`, `QUARANTINED`, `ATTESTED`).

## 5. Alternatives Considered
* **Global Whitelist Only:** Too restrictive; developers need project-specific flexibility.
* **Automatic Sandboxing (Docker):** High overhead and complexity for local development workflows. Rejected as a primary discovery-layer solution but supported as an execution-layer option.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Uses cryptographic signatures for "Tool Templates." Any change to the template (e.g., adding `; rm -rf /`) invalidates the attestation.
* **Observability:** All project-scoped discovery attempts and attestation results are logged to the audit trail.

## 7. Evolutionary Changelog
* **2026-03-01:** Initial Document Creation.
