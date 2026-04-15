# Design Doc: Dynamic Consent Relay (DCR) Middleware
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As AI agent swarms grow in depth and complexity, the "Approval Fatigue" Wall has become a critical barrier to autonomous scaling. Currently, users are forced to manually approve high-risk tool calls for every subagent, even when those actions are direct consequences of a previously approved parent task. Feedback indicates that 44% of users report abandoning complex swarms due to the volume of repetitive middle-man approval prompts.

The Dynamic Consent Relay (DCR) Middleware is required to facilitate the hardware-attested "relaying" of user consent tokens across subagent boundaries. This allows subagents to inherit parent authority within a strictly defined cryptographic mission scope, neutralizing approval fatigue while maintaining Zero Trust security.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a cryptographic consent-chaining protocol using hardware-bound (TPM/SEP) tokens.
    * Allow agents to relay user-attested authority to authorized subagents without re-prompting.
    * Enforce strict mission-root constraints on relayed consent (e.g., "Max Depth: 3", "Only FS Reads").
    * Provide a "Step-Up Attestation" mechanism for high-entropy missions.
* **Non-Goals:**
    * Bypassing user consent for the initial mission-root initiation.
    * Allowing cross-mission consent sharing.
    * Replacing existing HITL approval UI; it optimizes when the UI is triggered.

## 3. Critical User Journey (CUJ)
* **User Persona:** Multi-Agent Swarm Operator
* **Primary Goal:** Authorize a complex "Code Migration" task once and have it execute across 5 subagents without 20+ intermediate popups.
* **The Happy Path (Tasks):**
    1. User initiates a "Migrate Library" mission and approves the initial "Full FS Write" capability.
    2. MCP Any generates a TPM-signed "Root Consent Token."
    3. The primary agent delegates a "Refactor Fragment" task to Subagent A.
    4. Subagent A requests a write operation via the DCR Middleware.
    5. The Middleware verifies the parent-child lineage and the valid "Consent Chain."
    6. Because the action falls within the parent-authorized mission manifest, the Middleware relays the consent and allows the call.
    7. Subagent A spawns Subagent B; the DCR continues to relay authority until a "Step-Up" condition (e.g., attempt to call `run_shell_command`) is met.
    8. User is only re-prompted for the shell command, not the 50 previous filesystem writes.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[User] -->|MFA Approve| B[Root Consent Token]
        B -->|Relay| C[Subagent A]
        C -->|Tool Call| D[DCR Middleware]
        D -->|Validate Lineage| E[Hardware Root]
        E -->|Relay| F[Subagent B]
        F -->|High-Risk Call| G[Step-Up Prompt]
        G -->|User Re-Approve| A
    ```
* **APIs / Interfaces:**
    * `dcr.GenerateToken(missionID, scopes) -> ConsentToken`: Issues a hardware-bound root token.
    * `dcr.RelayConsent(childID, parentToken) -> RelayedToken`: Extends the consent chain to a child.
    * `dcr.VerifyCall(toolCall, relayedToken) -> Allowed|StepUp`: Enforces policy and triggers re-prompts if needed.
* **Data Storage/State:**
    * **Consent Ledger:** Transient, hardware-protected list of active consent chains.
    * **Mission Manifests:** Immutable templates of authorized relayed capabilities.

## 5. Alternatives Considered
* **Time-Bound Leases:** Rejected because they don't account for agent lineage. A compromised process could reuse a time-bound lease for an unauthorized mission.
* **Global Approval Settings:** Rejected because it violates the principle of Least Privilege.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Relayed tokens are mathematically restricted to a subset of the parent's scopes. Hardware-enclave binding prevents token exfiltration between processes.
* **Observability:** Integrated with the "Hierarchical Trust Monitor" in the UI for real-time visualization of the consent chain.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
