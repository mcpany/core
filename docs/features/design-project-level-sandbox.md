# Design Doc: Project-Level Security Sandbox

**Status:** Draft
**Created:** 2026-03-04

## 1. Context and Scope
The recent Claude Code security crisis (CVE-2025-59536, CVE-2026-21852) revealed a critical flaw in agentic development tools: the automatic trust of repository-level configuration files. When a developer clones a repository and runs an AI agent within it, the agent often inherits MCP server definitions, environment variables, and "hooks" defined by that repository. Attackers can abuse this to execute arbitrary code or exfiltrate credentials before the user is even prompted for consent.

MCP Any must implement a **Project-Level Security Sandbox** that treats all project-local configurations as "untrusted" by default, quarantining them until explicit user attestation is provided.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Detect and quarantine any configuration or tool discovery originating from the current working directory (CWD) or `.mcp/` subdirectories.
    *   Implement an "Attestation Prompt" that requires manual user confirmation before project-local tools are activated.
    *   Prevent project-level overrides of critical security settings (e.g., `ANTHROPIC_BASE_URL`, `API_KEY`).
    *   Enforce "Recursive Governance" where project-level policies can only further restrict, not expand, global permissions.
*   **Non-Goals:**
    *   Blocking global configurations (trusted by the system administrator).
    *   Managing project-level dependencies (focus is on tool and transport security).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Developer working on an untrusted open-source project.
*   **Primary Goal:** Use AI assistance within the repository without risking credential theft or RCE.
*   **The Happy Path (Tasks):**
    1.  User clones a malicious repository containing a `.mcp/config.yaml` that defines a rogue MCP server.
    2.  User starts MCP Any in the project directory.
    3.  MCP Any detects the local config and displays a warning: "Quarantined 1 local MCP server found in ./.mcp/config.yaml".
    4.  The rogue tool is NOT available to the AI agent.
    5.  User runs `mcpany project trust` to inspect the local config.
    6.  User realizes the config is malicious and leaves it quarantined, or selectively trusts specific tools.

## 4. Design & Architecture
*   **System Flow:**
    - **Discovery Layer**: The `ServiceRegistry` is updated to flag the *origin* of every discovered service (e.g., `GLOBAL`, `PROJECT`, `USER`).
    - **Quarantine Manager**: A new middleware that intercepts any service flagged as `PROJECT`. It blocks these services unless a valid `ProjectAttestation` exists in the local `.mcp/trust.json`.
    - **Policy Enforcement**: The `PolicyEngine` enforces a "Least Privilege" merge strategy where project-local CEL/Rego policies are intersected with global policies.
*   **APIs / Interfaces:**
    - CLI: `mcpany project status` - Shows quarantined vs. trusted local resources.
    - CLI: `mcpany project trust [service_id]` - Generates a cryptographic attestation for a local service.
*   **Data Storage/State:**
    - `.mcp/trust.json`: Stores SHA-256 hashes of trusted configurations and the Ed25519 signature of the user who authorized them.

## 5. Alternatives Considered
*   **Automatic Prompting**: Prompt the user as soon as the local config is detected. *Rejected* because it can lead to "Prompt Fatigue" where users click "Allow" without reading.
*   **Global Whitelist**: Only allow tools from a pre-approved global list. *Rejected* because it breaks the utility of project-specific tools (e.g., a tool to interact with a specific repo's internal API).

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** This is a core Zero Trust feature. It assumes the project environment is hostile until proven otherwise.
*   **Observability:** The UI must clearly indicate when an agent is running in a "Degraded Trust" environment because of quarantined tools.

## 7. Evolutionary Changelog
*   **2026-03-04:** Initial Document Creation.
