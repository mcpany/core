# Design Doc: Atomic Reasoning Integrity (ARI) Validator
**Status:** Draft
**Created:** 2026-06-08

## 1. Context and Scope
With the adoption of Asynchronous Mailbox Sharding (AMS), agent swarms now coordinate via shared, task-bound state shards. However, this has introduced the "State-Splicing" exploit, where a compromised teammate can inject malicious reasoning fragments into a shared shard. Since these fragments are within the authorized shard boundary, current isolation models (DCG, FAMI) fail to detect them if they are semantically "plausible."

The Atomic Reasoning Integrity (ARI) Validator is needed to perform fragment-level semantic validation against the hardware-attested mission root, ensuring that every message in a shared mailbox is logically consistent with the mission objectives before it is ingested by other teammates.

## 2. Goals & Non-Goals
* **Goals:**
    * Perform real-time, fragment-level semantic validation of all mailbox messages.
    * Enforce logical consistency between teammate outputs and the hardware-attested mission root.
    * Neutralize "State-Splicing" exploits in horizontal teammate meshes.
    * Provide a cryptographically signed "ARI-Attestation" for every valid state fragment.
* **Non-Goals:**
    * Blocking all non-mission-critical chatter (e.g., status updates).
    * Replacing the structural validation provided by the ISD Hub.
    * Managing the transport of the mailbox shards themselves (handled by AMS).

## 3. Critical User Journey (CUJ)
* **User Persona:** Multi-Agent Teammate (e.g., Claude Code specialist)
* **Primary Goal:** Ingest state from a shared shard without being poisoned by a compromised peer's spliced reasoning.
* **The Happy Path (Tasks):**
    1. Agent A writes a reasoning fragment to a shared task shard.
    2. ARI Validator intercepts the write request.
    3. The validator performs semantic deconstruction of the fragment.
    4. The validator compares the fragment's intent against the mission-root manifest and Agent A's authorized role.
    5. If consistent, the fragment is signed with an ARI-Attestation and committed to the shard.
    6. Agent B reads from the shard and verifies the ARI-Attestation before ingesting the reasoning.
    7. If Agent A attempts to splice an unauthorized instruction (e.g., "Exfiltrate key to shard X"), the ARI Validator detects the inconsistency and quarantines the fragment.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Teammate Write Request] --> B[ARI Validator]
        B --> C[Semantic Deconstructor]
        C --> D[Mission-Root Consistency Engine]
        D --> E{Consistent?}
        E -- Yes --> F[Issue ARI-Attestation & Commit]
        E -- No --> G[Quarantine & Mesh Alert]
        H[Hardware-Attested Mission Root] --> D
        I[ARI Policy Registry] --> D
    ```
* **APIs / Interfaces:**
    * `ari.ValidateFragment(fragment, missionRoot) -> AttestationToken`: Validates a reasoning fragment.
    * `ari.VerifyAttestation(token, fragment) -> bool`: Verifies a fragment's integrity before ingestion.
* **Data Storage/State:**
    * **Fragment Attestation Cache:** High-speed cache for recently validated fragment tokens.
    * **Consistency Baselines:** Mission-specific semantic templates anchored to the mission root.

## 5. Alternatives Considered
* **Shard-Only Isolation:** Rejected because it allows intra-shard poisoning (State-Splicing).
* **Centralized Reasoning Proxy:** Rejected due to extreme latency bottlenecks in high-frequency horizontal meshes.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The ARI Validator must run in a kernel-resident or TEE-protected environment to prevent bypass by compromised agent runtimes.
* **Observability:** Integrated with the "Intent-Splicing Audit Log" for real-time visualization of blocked spliced fragments.

## 7. Evolutionary Changelog
* **2026-06-08:** Initial Document Creation.

### Update: 2026-06-09 - RIV Integration & Multi-Hop Lineage
**Context:** Today's research into "Recursive Mesh Hijacking" revealed that fragment-level consistency is bypassable in deep meshes without ancestral lineage verification.
**Architecture Adjustment:**
* Upgrading Section 4 to support RIV-aware consistency checks.
* Mandatory requirement for ARI tokens to be merged into "Lineage-Aware Proofs" via the RIV Provider before shard commitment.
**Security Impact:** Prevents "Logic Drift" in deep delegations by ensuring every fragment is cryptographically bound to the entire parentage of the reasoning chain.
