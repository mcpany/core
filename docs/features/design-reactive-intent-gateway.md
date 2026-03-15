# Design Doc: Reactive Intent Gateway (RIG)
**Status:** Draft
**Created:** 2026-04-17

## 1. Context and Scope
As agents operate in dynamic environments, they often encounter scenarios where their initial "Mission Intent" is too restrictive (e.g., a tool requires access to a sibling directory not in the original allow-list). The Reactive Intent (RI) proposal from OpenClaw allows agents to request "Boundary Expansions." However, without a mediation layer, this creates a major "Intent Smuggling" vector.

The Reactive Intent Gateway (RIG) is a security middleware for MCP Any that intercepts, validates, and signs these expansion requests against the Root Mission Intent and local Zero-Trust policies.

## 2. Goals & Non-Goals
* **Goals:**
    * Intercept all "Boundary Expansion" requests from connected agents.
    * Validate expansion requests against cryptographically signed Root Mission Intents.
    * Enforce "Hardware-Attested Memory Pinning" to ensure constraints remain immutable.
    * Provide a standardized interface for User Attestation (HITL) for high-risk expansions.
* **Non-Goals:**
    * Automatically granting expansions without policy validation.
    * Modifying the agent's internal reasoning logic.

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Agent Orchestrator
* **Primary Goal:** Safely allow a subagent to expand its filesystem access to a specific sub-directory without exposing the entire host.
* **The Happy Path (Tasks):**
    1. Subagent encounters a permission error and generates a Reactive Intent (RI) request.
    2. MCP Any's RIG intercepts the RI request.
    3. RIG verifies the request against the Root Mission Intent.
    4. RIG checks if the expansion falls within the "Pre-Approved" policy for the current profile.
    5. RIG prompts the user for MFA attestation (via HITL) if the risk threshold is exceeded.
    6. Upon approval, RIG signs a new "Scoped Intent Token" and hands it back to the agent.

## 4. Design & Architecture
* **System Flow:**
    `Agent -> RI Request -> RIG Middleware -> Policy Engine -> [HITL] -> Signed Intent Token -> Agent`
* **APIs / Interfaces:**
    * `POST /v1/intent/expand`: Endpoint for agents to submit expansion requests.
    * `GetIntent(token)`: Service to retrieve and verify intent lineage.
* **Data Storage/State:**
    * Intent tokens are stored in the Shared KV Store (Blackboard) with "Parent-Bound" isolation.

## 5. Alternatives Considered
* **Implicit Expansion:** Rejected due to high risk of Intent Smuggling and sandbox escapes.
* **Static Intents Only:** Rejected as it leads to "Cognitive Lock" and high failure rates in dynamic tasks.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All expansion tokens must be hardware-bound (TPM) and session-specific.
* **Observability:** RIG logs all denied expansion attempts and attestation results for audit trails.

## 7. Evolutionary Changelog
* **2026-04-17:** Initial Document Creation.
