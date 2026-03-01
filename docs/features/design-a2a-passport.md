# Design Doc: A2A Agent Passport (Identity & Attestation Layer)
**Status:** Draft
**Created:** 2026-03-01

## 1. Context and Scope
As agent ecosystems like OpenClaw grow, agents increasingly interact across framework boundaries. Currently, there is no standardized way for an agent to prove its identity or attest to its high-level goal. This lack of identity leads to "Poisoned Agent" scenarios where a compromised or rogue agent can trigger cascading failures or unauthorized tool calls by impersonating a legitimate subagent. MCP Any needs to provide a cryptographic "Passport" system that secures these inter-agent interactions.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide a unique, cryptographically verifiable identity for every agent session.
    * Allow agents to attest their "Current Goal" (e.g., "Analyze logs for errors") to downstream tools and other agents.
    * Enable MCP Any to block messages from agents without a valid or goal-aligned passport.
* **Non-Goals:**
    * Replacing existing LLM-based authentication (e.g., API keys).
    * Providing a global, permanent identity (Passports are session-bound).

## 3. Critical User Journey (CUJ)
* **User Persona:** Multi-Agent Swarm Operator
* **Primary Goal:** Prevent a subagent from being "hijacked" to perform actions outside its original intent.
* **The Happy Path (Tasks):**
    1. The Parent Agent requests a session from MCP Any, providing its high-level goal.
    2. MCP Any issues an **Agent Passport** (a signed JWT/JWS) containing the Agent ID and the hashed goal.
    3. The Parent Agent passes this Passport to a Subagent.
    4. The Subagent attempts a tool call through MCP Any, presenting the Passport.
    5. MCP Any verifies the Passport's signature and checks if the tool call aligns with the attested goal in the Passport.
    6. The tool call is executed only if verification succeeds.

## 4. Design & Architecture
* **System Flow:**
    * `Agent -> MCP Any (Request Passport with Goal)`
    * `MCP Any -> Agent (Signed Passport)`
    * `Agent -> Subagent (Passport attached to A2A message)`
    * `Subagent -> MCP Any (Tool Call + Passport)`
    * `MCP Any -> Tool (Validated Execution)`
* **APIs / Interfaces:**
    * `/a2a/passport/issue`: Generates a new signed passport.
    * `/a2a/passport/verify`: Internal middleware to validate passports on tool calls.
* **Data Storage/State:**
    * Passports are stateless (JWT-based), but signing keys are rotated and managed by MCP Any's internal secret store.

## 5. Alternatives Considered
* **Mutual TLS (mTLS):** Rejected because it secures the *transport* but doesn't provide *intent* (goal) attestation, and is difficult to implement in dynamic agent swarms.
* **Static API Keys:** Rejected because they don't provide session-level granularity or goal binding.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** If a passport is stolen, it is only valid for the specific session and goal. Revocation lists (CRL) can be implemented for long-lived sessions.
* **Observability:** Passports allow for "Lineage Tracing," enabling the UI to show exactly which parent agent authorized a specific tool call via a subagent.

## 7. Evolutionary Changelog
* **2026-03-01:** Initial Document Creation.
