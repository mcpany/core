# Design Doc: Ephemeral Privilege Manager (EPM)
**Status:** Draft
**Created:** 2026-04-28

## 1. Context and Scope
The "BoryptGrab" security crisis has exposed a fundamental flaw in the "all-or-nothing" privilege model of modern AI agents. When a user grants an agent broad filesystem or system access to perform a complex task, that privilege often persists beyond the task's lifetime. Malicious subagents or Trojan-infected skills can then weaponize these persistent privileges to exfiltrate sensitive data (SSH keys, env vars) or install backdoors.

MCP Any needs to transition to an **Ephemeral Agency** model, where high-level privileges are only granted for the specific duration of a verified task and are automatically revoked.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a "Just-in-Time" (JIT) privilege escalation system.
    * Automatically revoke privileges after a configurable timeout or task completion signal.
    * Provide a cryptographic "Lease" for every high-risk tool call.
    * Integrate with the HITL Middleware for multi-factor authorization of escalation requests.
* **Non-Goals:**
    * Managing OS-level users or groups (uses existing service accounts).
    * Providing a full sandbox (relies on existing virtualization adapters).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious DevSecOps Engineer
* **Primary Goal:** Allow an agent to fix a production bug in a restricted directory without granting permanent root access.
* **The Happy Path (Tasks):**
    1. Agent identifies a bug in `/etc/nginx/conf.d/`.
    2. Agent requests a "Privilege Lease" for `fs:write:/etc/nginx/` from MCP Any.
    3. MCP Any evaluates the request against the "Semantic Risk Arbiter."
    4. User receives an MFA prompt on their mobile/CLI to approve the 5-minute lease.
    5. User approves; MCP Any generates a signed "Ephemeral Token."
    6. Agent uses the token to execute the `fs_write` tool.
    7. After 5 minutes, MCP Any invalidates the token, even if the agent is still running.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    sequenceDiagram
        Agent->>EPM: Request Lease (Scope, Timeout)
        EPM->>RiskArbiter: Evaluate Semantic Risk
        RiskArbiter-->>EPM: High Risk (MFA Required)
        EPM->>User: MFA Prompt (A2UI)
        User-->>EPM: Approve
        EPM->>TokenIssuer: Issue Ephemeral JWT
        TokenIssuer-->>Agent: Lease Token
        Agent->>ToolGateway: Call Tool(Token)
        ToolGateway->>EPM: Validate Token
        EPM-->>ToolGateway: Valid
        ToolGateway->>TargetTool: Execute
    ```
* **APIs / Interfaces:**
    * `POST /epm/request-lease`: Request a new privilege lease.
    * `GET /epm/validate-token`: Internal endpoint for the gateway to verify lease status.
* **Data Storage/State:**
    * In-memory "Lease Registry" (backed by SQLite Blackboard for persistence across restarts).
    * TTL-indexed entries for automatic expiration.

## 5. Alternatives Considered
* **Static Config:** Rejected because it doesn't solve the "persistence" problem.
* **OS-Level Sudoers:** Too complex to manage dynamically across different platforms (macOS, Linux, Windows).

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Tokens are bound to the specific agent session and origin.
* **Observability:** Every escalation and expiration is logged in the "Local Security Audit Log."

## 7. Evolutionary Changelog
* **2026-04-29:** Addressing "BoryptGrab" persistence by binding privilege leases to the ContextEngine's session lifecycle. Introduced "Lifecycle-Bound Revocation" to ensure high-risk capabilities are purged immediately upon subagent or task termination.
* **2026-04-28:** Initial Document Creation.
