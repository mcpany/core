# Design Doc: Autonomous PR Integrity Gate (APRIG)
**Status:** Draft
**Created:** 2026-05-28

## 1. Context and Scope
AI coding agents are capable of generating working software at unprecedented speeds, but they consistently introduce security vulnerabilities. A recent report indicates that 87% of agent-generated pull requests contain at least one security flaw. While AI-powered scanning helps, it is often bypassable or fails to capture logic-level vulnerabilities.

MCP Any needs to act as the final gate for any tool call that generates or modifies code (e.g., `git:create-pr`, `fs:write`). APRIG implements a multi-agent security quorum where a code change must be attested to by independent "Security Auditor" agents before it is committed to a repository.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Automatically intercept tool calls that result in code commits or pull requests.
    *   Orchestrate a "Security Quorum" of 2+ independent auditor agents.
    *   Provide a standardized "Attestation Receipt" that must be attached to the PR.
    *   Allow users to define custom "Audit Policies" (e.g., "Always check for hardcoded secrets").
*   **Non-Goals:**
    *   Replacing traditional CI/CD security scanners (APRIG is a pre-commit/pre-PR gate).
    *   Manually fixing the vulnerabilities (the auditors provide feedback for the primary agent to self-correct).

## 3. Critical User Journey (CUJ)
*   **User Persona:** DevOps Engineer
*   **Primary Goal:** Ensure that no AI-generated code is merged into the `main` branch without a verified security review.
*   **The Happy Path (Tasks):**
    1.  The Coding Agent generates a fix and calls `git:create-pr`.
    2.  MCP Any intercepts the call and triggers the **APRIG Workflow**.
    3.  Two independent Auditor Agents (e.g., one specialized in OWASP Top 10, another in project-specific patterns) receive the diff.
    4.  Auditors provide "Pass/Fail" signals and reasoning monologues.
    5.  If both Pass, the PR is created with an **APRIG Attestation Badge**.
    6.  If one Fails, the Coding Agent receives the feedback and must attempt a self-correction cycle before re-submitting to the gate.

## 4. Design & Architecture
*   **System Flow:**
    ```mermaid
    graph TD
        Agent[Coding Agent] -->|Tool Call: Create PR| Gate[APRIG Middleware]
        Gate -->|Request Audit| Auditor1[Security Auditor A]
        Gate -->|Request Audit| Auditor2[Security Auditor B]
        Auditor1 -->|Pass + Logic| Gate
        Auditor2 -->|Pass + Logic| Gate
        Gate -->|Consensus Reached| Git[Git MCP Server]
        Git -->|Success| Agent
    ```
*   **APIs / Interfaces:**
    *   `POST /v1/aprig/audit`: Triggers a quorum review for a provided diff.
    *   `GET /v1/aprig/status/:id`: Returns the current consensus status of an audit.
*   **Data Storage/State:** Audit logs and auditor monologues are stored in the Blackboard under the `aprig:audit_id` namespace.

## 5. Alternatives Considered
*   **Static Analysis Only:** Rejected because static tools miss the semantic context and "Intent" of the change.
*   **Mandatory HITL:** Rejected because it doesn't scale with high-frequency agent swarms. APRIG provides "Agentic HITL" with human oversight as an optional final step.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** Auditors must be cryptographically distinct from the primary agent to prevent collusion.
*   **Observability:** A dedicated "PR Integrity Dashboard" visualizes the pass/fail rates and common vulnerability patterns found by the swarm.

## 7. Evolutionary Changelog
*   **2026-05-28:** Initial Document Creation.
