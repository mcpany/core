# Design Doc: "Time-to-Trust" Progressive Permissions

**Status:** Draft
**Created:** 2026-03-02

## 1. Context and Scope
Autonomous agents are increasingly being given broad system permissions to improve their utility. However, this "Vibe-Code" approach circumvents system safety. As agents operate over longer periods, their reliability can be established. The "Time-to-Trust" framework (inspired by CSA 2026) formalizes this by escalating an agent's permissions based on time, successful tasks, and lack of security incidents.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Implement three trust tiers for agent sessions: `Probation`, `Junior`, and `Senior`.
    *   Define default capabilities for each tier (e.g., `Probation` = Read-only, No Internet).
    *   Automate tier escalation based on "Incident-Free Time" and "Manual Attestation".
    *   Provide an audit log for all permission escalations.
*   **Non-Goals:**
    *   Replacing the Policy Firewall (Time-to-Trust acts as a meta-policy layer).
    *   Managing user identities (focus is on agentic session trust).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Enterprise DevOps Security Lead.
*   **Primary Goal:** Allow developers to use autonomous coding agents without risking production data or credentials.
*   **The Happy Path (Tasks):**
    1.  New agent session starts in `Probation` mode.
    2.  Agent performs file reads and analysis for 4 hours without violating any `Safe-by-Default` policies.
    3.  System automatically notifies the user that the agent is eligible for `Junior` status.
    4.  User approves escalation; agent can now perform local file writes in specified directories.

## 4. Design & Architecture
*   **System Flow:**
    - **Trust Manager**: Tracks session duration, tool call frequency, and policy violations.
    - **Escalation Engine**: Evaluates a session's trust level against configured thresholds (e.g., `Senior` requires 30 days of activity and 0 incidents).
    - **Policy Integration**: The `Policy Firewall` injects the current `TrustLevel` into its evaluation context.
*   **APIs / Interfaces:**
    - `GET /sessions/{id}/trust-status`: Returns current tier and time until next escalation.
    - `POST /sessions/{id}/attest-trust`: Manual user override to promote/demote trust.
*   **Data Storage/State:** Persistent session metadata stored in the embedded SQLite database.

## 5. Alternatives Considered
*   **Static RBAC**: assigning fixed roles to agents. *Rejected* as it doesn't account for the dynamic reliability of agentic behavior.
*   **Model-Based Trust**: letting the LLM decide if it's "ready" for more access. *Rejected* due to prompt injection risks.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** If an incident (policy violation) occurs, the trust level is immediately reset to `Probation`.
*   **Observability:** The UI must display the agent's current "Probation Timer."

## 7. Evolutionary Changelog
*   **2026-03-02:** Initial Document Creation.
