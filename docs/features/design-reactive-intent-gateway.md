# Design Doc: Reactive Intent Arbitration Hub
**Status:** Draft
**Created:** 2026-04-17

## 1. Context and Scope
As agent swarms utilize "Reactive Intent" (RI) to dynamically request boundary expansions, a new attack vector called "Intent Smuggling" has emerged. Malicious subagents can embed unauthorized secondary goals within a legitimate expansion request. MCP Any needs an "Arbitration Hub" that recursively deconstructs these requests and validates them against the cryptographically signed Root Mission Intent.

## 2. Goals & Non-Goals
* **Goals:**
    * Recursively deconstruct "Boundary Expansion" requests into atomic sub-goals.
    * Validate each sub-goal against the "Root Mission Intent" and current security policy.
    * Provide a cryptographic "Arbitration Proof" for approved expansions.
    * Integrate with the UACO v2.2 Intent Barrier middleware to prevent race conditions during expansion.
* **Non-Goals:**
    * Automatically generating the expansion request (handled by the Agent Framework).
    * Modifying the Root Mission Intent (requires explicit user/parent re-attestation).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Swarm Architect
* **Primary Goal:** Prevent a specialized "Code Formatter" subagent from smuggling a "File Exfiltration" intent during a legitimate "Write Access" expansion request.
* **The Happy Path (Tasks):**
    1. Subagent encounters a "Permission Denied" error when trying to write to a project sub-directory.
    2. Subagent generates a Reactive Intent expansion request for `fs:write`.
    3. The Arbitration Hub intercepts the request.
    4. The Hub deconstructs the request and finds a hidden `net:outbound` sub-goal.
    5. The Hub validates the sub-goals against the Root Intent ("Refactor Code") and identifies the network goal as a smuggled intent.
    6. The Hub blocks the request and alerts the parent agent/user.
    7. If the request was clean, the Hub would issue a signed "Arbitration Proof" to the A2A Messaging Hub.

## 4. Design & Architecture
* **System Flow:**
    `[Expansion Request] -> [Deconstructor] -> [Intent Matcher] -> [Policy Validator] -> [Arbitration Proof]`
* **APIs / Interfaces:**
    * `ArbitrationHub`: `ValidateExpansion(request RIRequest, rootIntent Intent) (ArbitrationProof, error)`
    * `Deconstructor`: `Decompose(request RIRequest) []AtomicGoal`
* **Data Storage/State:**
    * Caches verified intent trees and arbitration proofs in the Blackboard.

## 5. Alternatives Considered
* **Flat Intent Validation**: Rejected as it cannot detect smuggled goals hidden within complex request objects.
* **Manual Approval for All Expansions**: Rejected due to "Approval Fatigue" in high-frequency swarms.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The Arbitration Hub must use hardware-bound signatures for all issued proofs.
* **Observability:** Smuggling attempts are logged as high-priority security alerts in the RIG Dashboard.

## 7. Evolutionary Changelog
* **2026-04-17:** Initial Document Creation.
