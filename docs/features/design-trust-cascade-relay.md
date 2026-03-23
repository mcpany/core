# Design Doc: Trust Cascade Relay (TCR)
**Status:** Draft
**Created:** 2026-06-28

## 1. Context and Scope
Deep agent swarms (e.g., A -> B -> C -> D) are currently bottlenecked by "Cognitive Stall." Each hop in the delegation chain requires a full, hardware-attested A2A handshake (mTLS + TPM signature verification), which can add 200ms-500ms of latency. In high-frequency reasoning loops, this latency often leads to session timeouts or "Handshake Fatigue."

The TCR implements the "Trust Cascade Protocol" (TCP). It allows a verified mission root or a parent agent to issue ephemeral, session-bound "Trust Fragments" to its descendants. These fragments carry a cryptographically signed proof of the parent's attestation, allowing children to prove their authority to sub-specialists without repeating a full hardware handshake.

## 2. Goals & Non-Goals
* **Goals:**
    * Reduce inter-agent delegation latency by 80%+.
    * Provide a hardware-locked lineage for trust propagation.
    * Support ephemeral, mission-bound trust fragments.
    * Implement a "Fragment Replay Guard" to prevent unauthorized reuse.
* **Non-Goals:**
    * Replacing the initial root-mission attestation.
    * Bypassing tool-level security (trust fragments only authorize the *delegation*, not the tool call itself).

## 3. Critical User Journey (CUJ)
* **User Persona:** High-Frequency Swarm Orchestrator
* **Primary Goal:** Execute a 5-hop delegation chain in under 1 second.
* **The Happy Path (Tasks):**
    1. Mission Root Agent completes a full hardware handshake with TCR.
    2. Root Agent spawns Specialist A and issues a "Trust Fragment" via TCR.
    3. Specialist A delegates to Specialist B, passing the fragment.
    4. Specialist B validates the fragment locally against TCR's public key (sub-10ms).
    5. The chain continues to Specialist E with minimal latency.
    6. All tool calls along the chain are still attributed to the verified Mission Root.

## 4. Design & Architecture
* **System Flow:**
    [Mission Root] --(TPM Sign)--> [TCR] --(Issue Fragment)--> [Subagent A]
    [Subagent A] --(Forward Fragment)--> [Subagent B]
    [Subagent B] --(Validate Fragment)--> [TCR Local Cache]
* **APIs / Interfaces:**
    * `POST /v1/tcr/issue`: Exchange a full hardware signature for a session-bound trust fragment.
    * `POST /v1/tcr/verify`: Validates a fragment and its lineage.
* **Data Storage/State:**
    Fragments are ephemeral and stored in memory-mapped buffers. Fragment nonces are persisted in the "Fragment Replay Guard" (SQLite).

## 5. Alternatives Considered
* **Persistent mTLS Tunnels:** Rejected because agents are often short-lived processes, and tunnel overhead is high.
* **Shared Session Keys:** Rejected as it lacks the hardware-bound lineage required for Zero-Trust.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Fragments expire automatically upon mission completion or parent termination.
* **Observability:** TCR dashboard visualizes the "Trust Tree" and highlights any replay attempts.

## 7. Evolutionary Changelog
* **2026-06-28:** Initial Document Creation.
