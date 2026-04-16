# Design Doc: Optimistic Context Hub (OCH)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
High-density P2P meshes are increasingly suffering from "Semantic Deadlocks" (as observed in OpenClaw v3.6.2), where agents stall their reasoning loops while waiting for mutual context attestation from peers. The "Universal Agent Bus" must maintain high velocity without compromising its Zero-Trust security model. OCH solves this by introducing a speculative coordination pattern.

## 2. Goals & Non-Goals
* **Goals:**
    * Enable speculatively ingestion of peer context fragments to eliminate coordination stalls.
    * Maintain a "Probabilistic Buffer" for un-attested state fragments.
    * Trigger automated state rollbacks if background hardware-attested quorums fail.
    * Provide "Speculative Tags" to agents so they can weigh the confidence of their current reasoning.
* **Non-Goals:**
    * Bypassing security quorums (quorums still run in the background).
    * Persisting un-attested data to the long-term mission-root blackboard.

## 3. Critical User Journey (CUJ)
* **User Persona:** Multi-Agent Swarm Orchestrator
* **Primary Goal:** Resolve a circular dependency between a "Code Generator" and a "Linter" without a 5s+ semantic deadlock.
* **The Happy Path (Tasks):**
    1. The Code Generator produces a fragment and broadcasts it to the Linter.
    2. OCH intercepts the fragment and places it in the Linter's "Probabilistic Buffer" with a "Speculative" tag.
    3. The Linter begins reasoning against the fragment immediately.
    4. Simultaneously, the OCH initiates a hardware-attested discovery quorum in the background.
    5. The quorum verifies the fragment's integrity within 200ms.
    6. OCH promotes the fragment to "Attested" status; the Linter's reasoning proceeds to the commit phase.
    7. Swarm velocity is maintained despite the "Attestation Tax."

## 4. Design & Architecture
* **System Flow:**
    `[Peer A Fragment] -> [OCH] -> [Probabilistic Buffer] -> [Peer B Reasoning]`
    `                     |                                     ^`
    `                     +-> [Background Quorum] -> [Verification Signal] --+`
* **APIs / Interfaces:**
    * `och.v1.IngestSpeculativeFragment(content Fragment) (FragmentID, error)`
    * `och.v1.GetVerificationStatus(fragmentID string) (StatusEnum)`
* **Data Storage/State:**
    * Speculative fragments are held in volatile, isolated RAM segments.
    * Transition to RAMS-compliant storage occurs only upon verification.

## 5. Alternatives Considered
* **Synchronous Quorums:** Rejected due to "Semantic Deadlock" risks in large meshes.
* **Trust-on-First-Use (TOFU):** Rejected as it violates Zero-Trust principles for autonomous agents.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** OCH must ensure that no "Speculative" data can trigger a side-effect (e.g., a tool call with host access) until it is promoted to "Attested."
* **Observability:** "Rollback Rates" and "Speculative Latency Gains" are tracked in the OCH performance dashboard.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
