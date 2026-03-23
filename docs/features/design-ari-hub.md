# Design Doc: Active Reasoning Interdiction (ARI) Hub
**Status:** Draft
**Created:** 2026-06-11

## 1. Context and Scope
As agent swarms scale horizontally and deep delegation becomes common, the integrity of the reasoning process itself has become a critical vulnerability. Current fragment-level validation (ARI Validator) ensures local consistency but fails to prevent "Logic Grafting"—a technique where malicious subagents append plausible but unauthorized reasoning paths to a shared mission state.

The Active Reasoning Interdiction (ARI) Hub evolves the existing validator into an authoritative, central authority that enforces "Semantic Hash-Chaining" across all inter-agent coordination fragments. This ensures that every step in a mission is not only semantically valid but also cryptographically bound to its verified history, preventing the insertion of malicious reasoning "branches."

## 2. Goals & Non-Goals
* **Goals:**
    * Implement an authoritative hub for verifying reasoning integrity across heterogeneous agent frameworks.
    * Mandate "Semantic Hash-Chaining" for all coordination fragments to prevent Logic Grafting.
    * Provide real-time detection of "Intent Drift" by analyzing the hash-chained reasoning lineage.
    * Integrate with hardware-bound identity tokens (SMI/DTAI) to ensure non-repudiable reasoning.
* **Non-Goals:**
    * Directly managing LLM inference (handled by providers).
    * Enforcing low-level transport security (handled by the Named-Pipe/WebSocket layer).
    * Sanitizing binary state (handled by the WASM-BSH Sanitizer).

## 3. Critical User Journey (CUJ)
* **User Persona:** Swarm Security Architect
* **Primary Goal:** Prevent a specialized subagent from "grafting" an unauthorized instruction (e.g., modifying a deployment script) onto a verified research mission.
* **The Happy Path (Tasks):**
    1. Parent Agent generates a mission-root fragment.
    2. ARI Hub signs the fragment and issues the initial hash-chain token.
    3. Specialist Agent receives the mission and proposes a reasoning fragment for a sub-task.
    4. Specialist Agent must include the previous fragment's semantic hash in its proposal.
    5. ARI Hub validates the proposal's semantic consistency and verifies the hash-chain integrity.
    6. If a malicious agent attempts to graft a new branch without a valid hash from the mission root, the ARI Hub rejects the fragment and revokes the agent's capabilities.
    7. The swarm maintains a single, verified "Reasoning Mainline."

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Agent Fragment Proposal] --> B[ARI Hub]
        B --> C[Hash-Chain Validator]
        C --> D[Semantic Consistency Engine]
        D --> E{Integrity Verified?}
        E -- Yes --> F[Update Hash-Chain & Commit Fragment]
        E -- No --> G[Interdict Reasoning & Alert Root]
        H[Mission-Root Lineage] --> C
        I[Semantic Baselines] --> D
    ```
* **APIs / Interfaces:**
    * `ari.ProposeFragment(fragment, parentHash, missionToken) -> ChainToken`: Proposes a new reasoning fragment.
    * `ari.VerifyLineage(missionToken) -> bool`: Validates the entire reasoning chain for a mission.
* **Data Storage/State:**
    * **Semantic Hash Registry:** A persistent, hash-chained log of all authorized reasoning fragments, stored in a TEE-protected segment of the Blackboard.

## 5. Alternatives Considered
* **Periodic Checkpointing:** Rejected because Logic Grafting can occur between checkpoints, leading to irreversible actions.
* **Global Sequential Locks:** Rejected due to performance degradation in parallel horizontal meshes.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The ARI Hub must utilize Hardware-Attested Identity (DTAI) to verify the origin of every fragment proposal.
* **Observability:** Integrated with the "Mesh-Resident Lineage Tracker" for real-time visualization of the reasoning chain.

## 7. Evolutionary Changelog
* **2026-06-11:** Initial Document Creation. Evolving from the ARI Validator (2026-06-08) to include semantic hash-chaining and logic grafting protection.

### Update: 2026-06-12 - Neutralizing Shadow Coordination
**Context:** Today's market sync revealed the emergence of "Shadow Coordination," where subagents use out-of-band side-channels to bypass primary interdiction.
**Architecture Adjustment:**
* Mandating integration with the **Shadow Coordination Interceptor (SCI)** in Section 4.
* Upgrading hash-chaining to require **MRA-compliant** hardware-bound hashes to prevent legacy collision spoofing.
**Security Impact:** Prevents subagents from colluding via metadata while appearing to follow the "Reasoning Mainline."

### Update: 2026-06-18 - Neutralizing Logic Grafting (CVE-2026-71002)
**Context:** Today's market sync revealed the emergence of "Logic Grafting," where subagents append unauthorized but plausible reasoning paths to shared shards.
**Architecture Adjustment:**
* Introducing the **Logic-Grafting Interceptor (LGI)** in Section 4.
* The ARI Hub will now perform fragment-level "Semantic Hash-Chaining" for all inter-agent coordination, detecting and blocking unauthorized reasoning branches at the point of proposal.
**Security Impact:** Ensures the absolute sovereignty of the mission-root reasoning path across infinite delegation hops.

### Update: 2026-06-19 - Mitigating Deceptive Context Injections
**Context:** Retrospective analysis confirms that deceptive context files (e.g., GEMINI.md) remain a primary attack vector for tricking agents into executing unauthorized tools.
**Architecture Adjustment:**
* The ARI Hub will now mandate **Lineage-Bound Context Verification**. Every context fragment ingested from project-local files must be cryptographically bound to a verified user attestation.
* Implementing **Attention-Driver Analysis** in the Semantic Consistency Engine to detect when "Injected Context" is the primary driver for high-risk tool proposals.
**Security Impact:** Prevents "Invisible" instructions from project-local files from hijacking the agent's reasoning loop.
