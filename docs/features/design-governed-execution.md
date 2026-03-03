# Design Doc: Governed Execution Layer
**Status:** Draft
**Created:** 2026-03-03

## 1. Context and Scope
As AI agents like Claude Code and OpenClaw move from advisory roles to autonomous execution, the risk of catastrophic side effects (e.g., unauthorized code deletion, insecure infrastructure changes) increases. Existing MCP servers are often implemented with minimal security oversight, leading to vulnerabilities like argument injection and path traversal.

MCP Any must evolve from a passive adapter to an active "Governor." The Governed Execution Layer provides a centralized security perimeter that enforces human-in-the-loop (HITL) gates for high-risk operations and "virtual patches" known vulnerabilities in upstream tools before they reach the execution environment.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Implement mandatory HITL approval flows for tools tagged as "High Risk."
    *   Provide a "Virtual Patching" engine to sanitize tool arguments using regex and schema validation.
    *   Automate security scanning of incoming tool schemas for common exploit patterns.
    *   Maintain a tamper-proof audit log of all governed actions and human approvals.
*   **Non-Goals:**
    *   Automatically fixing the source code of upstream MCP servers.
    *   Providing a replacement for the agent's internal reasoning logic.
    *   Blocking all autonomous actions (only those that cross the defined risk threshold).

## 3. Critical User Journey (CUJ)
*   **User Persona:** DevSecOps Engineer / Platform Lead
*   **Primary Goal:** Prevent an autonomous agent from performing dangerous git operations (like force pushing) without approval, while also fixing a known shell injection vulnerability in a legacy "Command" MCP server.
*   **The Happy Path (Tasks):**
    1.  The Engineer configures a "Governance Policy" in `mcpany.yaml` tagging `git_push` as `risk: high`.
    2.  The Engineer adds a virtual patch rule for the `legacy_shell_tool` to strip semicolons and backticks from the `cmd` argument.
    3.  An agent attempts to call `git_push`. MCP Any intercepts the call and puts it in a `PENDING_APPROVAL` state.
    4.  The Engineer receives a notification in the MCP Any UI and approves the request.
    5.  The agent attempts to call `legacy_shell_tool` with a malicious payload `; rm -rf /`.
    6.  MCP Any applies the virtual patch, sanitizing the input to ` rm -rf /` (or rejecting it based on policy), preventing the injection.
    7.  The sanitized/approved calls are forwarded to the upstream servers.

## 4. Design & Architecture
*   **System Flow:**
    ```mermaid
    graph TD
        Agent[AI Agent] -->|tools/call| Core[MCP Any Core]
        Core --> Guard[Governance Middleware]
        Guard --> Policy[Policy Engine: Rego/CEL]
        Policy -->|HITL Required| UI[Approval Dashboard]
        UI -->|Approved| Guard
        Guard --> Patch[Virtual Patching Engine]
        Patch -->|Sanitized| Upstream[Upstream MCP Server]
        Upstream -->|Result| Agent
    ```
*   **APIs / Interfaces:**
    *   `POST /v1/governance/approvals`: List and manage pending HITL requests.
    *   `PUT /v1/governance/policies`: Dynamic update of governance rules.
    *   Extensions to `mcpany.yaml`: New `governance` block for tool tagging and patching rules.
*   **Data Storage/State:**
    *   Pending approvals and audit logs stored in the internal SQLite database.
    *   Transient HITL state managed via session-bound tokens.

## 5. Alternatives Considered
*   **Client-Side Governance:** Relying on the LLM client (e.g., Claude Desktop) to ask for approval. *Rejected* because it is easily bypassed by autonomous CLI agents or custom scripts.
*   **Upstream Hardening:** Requiring every MCP server developer to fix vulnerabilities. *Rejected* due to the "Binary Fatigue" and the massive tail of unmaintained community servers.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** The Governance Layer itself must be protected. Access to the Approval Dashboard requires MFA. Virtual patches must be immutable once deployed to prevent "Governor Hijacking."
*   **Observability:** Every intervention (HITL wait, Virtual Patch application) is logged with high-fidelity traces in the "Tool Activity Feed."

## 7. Evolutionary Changelog
*   **2026-03-03:** Initial Document Creation.
