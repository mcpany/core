# Design Doc: Delegated Trust Envelope (DTE) Broker
**Status:** Draft
**Created:** 2026-04-18

## 1. Context and Scope
As agent swarms scale, "Approval Fatigue" becomes a critical failure point. Users cannot manually approve every tool call in a deep chain. However, implicit trust leads to "Intent Smuggling." The DTE Broker provides a cryptographic middle ground: parent agents (or humans) can issue a "Trust Envelope" that delegates a specific, scoped subset of their authority to a subagent for a finite duration.

## 2. Goals & Non-Goals
* **Goals:**
    * Enable cryptographically bound permission delegation (Parent -> Subagent).
    * Enforce strict temporal and mission-scoped boundaries on DTEs.
    * Provide a universal bridge for DTEs across OpenClaw, AutoGen, and Gemini frameworks.
    * Automate subagent tool calls that fall within a valid DTE.
* **Non-Goals:**
    * Create a new global identity system (uses existing UAB/A2A identities).
    * Replace the Policy Firewall (DTEs are *additional* constraints).

## 3. Critical User Journey (CUJ)
* **User Persona:** Swarm Orchestrator
* **Primary Goal:** Delegate "Read-Only" access to a specific sub-directory to a research subagent for 10 minutes.
* **The Happy Path (Tasks):**
    1. Parent agent requests a DTE for `Subagent-Research`.
    2. Parent signs the envelope with `scope: fs:read:/docs/research` and `expiry: +10m`.
    3. DTE Broker validates the parent's authority to delegate these permissions.
    4. `Subagent-Research` presents the DTE to MCP Any when calling a tool.
    5. MCP Any verifies the cryptographic signature and allows the call without prompting the user.

## 4. Design & Architecture
* **System Flow:**
    `Parent Agent -> DTE Broker (Issue) -> Subagent -> MCP Any (Verify) -> Tool`
* **APIs / Interfaces:**
    * `IssueEnvelope(delegator, delegatee, scope, ttl) -> DTE`
    * `VerifyEnvelope(DTE, toolCall) -> boolean`
* **Data Storage/State:**
    * Ephemeral storage for active DTEs.
    * Revocation list (CRL) for compromised or prematurely terminated envelopes.

## 5. Alternatives Considered
* **Session Tokens:** Using simple bearer tokens. Rejected because they lack the cryptographic binding and lineage required for Zero-Trust swarms.
* **Dynamic Policy Updates:** Updating the Policy Firewall in real-time. Rejected as too complex to manage and audit at high frequency.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** DTEs are hardware-bound (TPM/SEP) whenever possible to prevent token theft.
* **Observability:** Visualization of the "Delegation Tree" in the UI to show how trust is flowing through the swarm.

## 7. Evolutionary Changelog
* **2026-04-18:** Initial Document Creation.
