# Design Doc: Recursive Intent Delegation (RID) Validator
**Status:** Draft
**Created:** 2026-03-26

## 1. Context and Scope
As agent swarms grow deeper, the risk of "Intent Ghosting" (where subagents lose track of the primary goal) and "Intent Hijacking" (where a compromised subagent redirects the swarm) increases. UACO v1.8 introduces Recursive Intent Delegation (RID) to provide a cryptographic framework for enforcing goal stability across recursive agent calls.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement validation for UACO v1.8 RID headers.
    * Enforce `max_depth` limits on recursive agent delegations.
    * Validate "Mutation Policies" to ensure subagents only modify intent within approved boundaries.
    * Integrate hardware-bound attestation (TPM/Secure Enclave) for root intent signing.
* **Non-Goals:**
    * Automatically "fix" drifted intents (RID is for validation/blocking).
    * Replace the Policy Firewall (RID is a specific layer of the firewall).

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise Swarm Architect
* **Primary Goal:** Ensure that a customer support swarm never deviates from its "Support Only" intent, even if a sub-sub-agent is compromised.
* **The Happy Path (Tasks):**
    1. Parent agent creates a Task Card with a signed RID certificate.
    2. Subagent receives the task and attempts to delegate a sub-task.
    3. MCP Any intercepts the delegation request.
    4. Validator verifies the RID signature against the Hardware Attestor.
    5. Validator checks that the new intent mutation complies with the parent's policy.
    6. Delegation is approved and routed to the next agent.

## 4. Design & Architecture
* **System Flow:**
    `[Agent Request] -> [RID Validator] -> [Hardware Attestor (TPM)] -> [Success/Failure]`
* **APIs / Interfaces:**
    * Internal `ValidateIntent(request, certificate)` interface.
    * UACO-compliant header parsing for `X-RID-Certificate`.
* **Data Storage/State:**
    * Cache of verified public keys from the Hardware-Bound Intent Attestor.

## 5. Alternatives Considered
* **Purely Prompt-Based Constraints:** Rejected because they are susceptible to injection and provide no cryptographic guarantees.
* **Centralized Intent Manager:** Rejected to avoid a single point of failure and maintain the decentralized nature of UACO swarms.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The validator itself must run in a high-integrity environment.
* **Observability:** Log all intent mutation attempts, highlighting those that were blocked.

## 7. Evolutionary Changelog
* **2026-03-26:** Initial Document Creation.
