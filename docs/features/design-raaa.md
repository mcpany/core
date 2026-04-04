# Design Doc: Risk-Aware Adaptive Attestation (RAAA)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
High-security agent swarms currently require frequent hardware-attested approvals for inter-agent coordination and tool execution. This "Zero-Trust" rigor has led to significant "Validation Fatigue" among users, where constant MFA prompts disrupt the workflow and lead to "Approval Burnout," potentially causing users to approve requests without proper scrutiny.

MCP Any needs to solve this by introducing Risk-Aware Adaptive Attestation (RAAA). RAAA is an intelligent governance layer that dynamically balances security and usability. By calculating real-time "Risk Scores" for every requested action, the system can automatically utilize background trust leases for low-risk, repetitive tasks while reserving high-strength hardware attestation (MFA) for high-impact or anomalous operations.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a dynamic Risk Engine that evaluates tool impact, data sensitivity, and agent historical behavior.
    * Dynamically adjust attestation strength requirements (e.g., Background Lease, Session Ticket, or Full MFA).
    * Reduce the frequency of user-facing attestation prompts by at least 60% for trusted missions.
    * Maintain hardware-bound lineage for all background-approved actions.
* **Non-Goals:**
    * Replacing Zero-Trust architecture with a "probabilistic" model.
    * Bypassing user-defined "Hard Deny" or "Mandatory MFA" rules for specific tools.
    * Automating high-risk actions (e.g., credential deletion) without human intervention.

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise Swarm Security Administrator
* **Primary Goal:** Enable high-speed autonomous coding swarms without compromising host security or burning out the human-in-the-loop.
* **The Happy Path (Tasks):**
    1.  The user defines a mission-root manifest with base trust levels for a coding team.
    2.  An agent specialist requests a `read_file` tool call for a standard library.
    3.  RAAA Middleware calculates a "Low Risk" score based on the mission scope and tool profile.
    4.  RAAA verifies an active "Fast-Path" trust lease and authorizes the call in the background.
    5.  The same agent later requests a `write_file` tool call to a restricted configuration directory.
    6.  RAAA calculates a "High Risk" score and immediately triggers a hardware-bound MFA prompt to the user.
    7.  The user, not being fatigued by low-risk prompts, performs a rigorous review and approves the specific mutation.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Agent Request] --> B[RAAA Middleware]
        B --> C{Risk Engine}
        C -->|Low Risk| D[Verify FPIR Trust Lease]
        C -->|High Risk| E[Trigger Hardware MFA]
        D -->|Valid| F[Authorize Tool Execution]
        D -->|Expired/Missing| E
        E -->|Approved| F
        E -->|Denied| G[Log & Terminate Branch]
        F --> H[Update Agent Reputation/History]
    ```
* **APIs / Interfaces:**
    * `X-MCP-Risk-Level`: Header injected into tool calls by the Risk Engine.
    * `/v1/governance/risk-policy`: Endpoint for defining risk thresholds per mission-root.
    * `RiskProfile` object in UAB task cards.
* **Data Storage/State:**
    * **Risk Metrics:** Stored in the intent-sealed Blackboard shards, tracking per-agent and per-tool confidence scores.
    * **Trust Leases:** Managed by the FPIR (Fast-Path Identity Resumption) provider in kernel-bound memory.

## 5. Alternatives Considered
* **Static Thresholds:** Pre-defined "trusted" tools. Rejected because risk is context-dependent (e.g., `git push` to a feature branch vs. `main`).
* **AI-Only Approval:** Using a "Supervisor Agent" to approve subagent actions. Rejected because it violates the Zero-Trust mandate for human-attested mission-root sovereignty.
* **Global Trust Tickets:** One MFA prompt for a 4-hour session. Rejected due to the risk of "Session Hijacking" where a compromised agent could perform high-risk actions once the ticket is issued.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All risk scores and background authorizations must be cryptographically bound to the mission-root hardware ID (TPM/SEP). RAAA must "Fail Secure" to full MFA if the Risk Engine is unreachable or returns an ambiguous score.
* **Observability:** Audit logs will record the "Reasoning Path" for every RAAA decision, allowing admins to visualize why a specific call was downgraded to a background lease or upgraded to MFA.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
