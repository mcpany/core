# Design Doc: Chain-of-Thought Lineage Provider (CTLP)
**Status:** Draft
**Created:** 2026-06-18

## 1. Context and Scope
As agent swarms become deeper and more autonomous, the "Black Box" reasoning problem has become a critical security and audit liability. Mission-root agents often delegate tasks to specialized subagents without a verifiable way to audit the "How" behind subagent conclusions. The CTLP aims to provide hardware-attested, non-repudiable lineage for every reasoning fragment in a swarm.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide hardware-attested reasoning traces for all inter-agent coordination.
    * Ensure non-repudiable audit trails for every sub-instruction.
    * Enable real-time lineage verification by parent agents.
    * Support mission-root sovereignty across infinite delegation hops.
* **Non-Goals:**
    * Performing semantic validation of the reasoning content (handled by ARI/AID).
    * Providing long-term archival storage for all reasoning traces.

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise Security Auditor
* **Primary Goal:** Audit a high-risk tool call made by a 4th-level subagent and verify its complete reasoning lineage back to the mission root.
* **The Happy Path (Tasks):**
    1. A subagent makes a high-risk tool call.
    2. The CTLP intercepts the request and retrieves the hardware-attested reasoning chain.
    3. The Auditor uses the CTLP dashboard to visualize the "Chain of Thought" lineage.
    4. The CTLP verifies the cryptographic signatures of every reasoning step against the Mission-Root TPM.
    5. The Auditor confirms that the tool call was derived from an authorized mission intent.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        MR[Mission Root] -->|Spawn| S1[Subagent 1]
        S1 -->|Reasoning Fragment 1| CTLP[Lineage Provider]
        CTLP -->|Attest & Hash| TPM[Hardware TPM]
        S1 -->|Spawn| S2[Subagent 2]
        S2 -->|Reasoning Fragment 2| CTLP
        CTLP -->|Link to Fragment 1| TPM
        S2 -->|Tool Call| Gateway[Secure Gateway]
        Gateway -->|Verify Lineage| CTLP
        CTLP -->|Validate Chain| TPM
    ```
* **APIs / Interfaces:**
    * `POST /v1/lineage/fragment/add`: Append a new reasoning fragment to the attested lineage.
    * `GET /v1/lineage/trace/{session_id}`: Retrieve the full reasoning trace for a mission session.
    * `POST /v1/lineage/verify`: Verify the cryptographic integrity of a reasoning chain.
* **Data Storage/State:**
    * Reasoning traces are stored as a hash-chained sequence of hardware-attested fragments in a session-bound buffer.

## 5. Alternatives Considered
* **Log-Based Lineage:** Rejected as logs are easily tampered with and lack hardware-bound non-repudiation.
* **Parent-only Tracking:** Rejected as it fails to capture the full reasoning path in deep agent swarms.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All lineage fragments are signed by hardware enclaves. The hash-chain ensures that fragments cannot be re-ordered or omitted.
* **Observability:** Integrated with the Mesh-Resident Lineage Tracker for real-time visualization of reasoning paths.

## 7. Evolutionary Changelog
* **2026-06-18:** Initial Document Creation.
