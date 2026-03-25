# Design Doc: Active Reasoning Interdiction (ARI) Hub
**Status:** Draft
**Created:** 2026-06-11

## 1. Context and Scope
As agent swarms scale horizontally and deep delegation becomes common, the integrity of the reasoning process itself has become a critical vulnerability. Current fragment-level validation (ARI Validator) ensures local consistency but fails to prevent "Logic Grafting" — a technique where malicious subagents append plausible but unauthorized reasoning paths to a shared mission state.

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
* **Primary Goal:** Prevent a specialized subagent from "grafting" an unauthorized instruction onto a verified research mission.
* **The Happy Path (Tasks):**
    1. Parent Agent generates a mission-root fragment.
    2. ARI Hub signs the fragment and issues the initial hash-chain token.
    3. Specialist Agent receives the mission and proposes a reasoning fragment for a sub-task.
    4. Specialist Agent must include the previous fragment's semantic hash in its proposal.
    5. ARI Hub validates the proposal's semantic consistency and verifies the hash-chain integrity.
    6. If a malicious agent attempts to graft a new branch without a valid hash from the mission root, the ARI Hub rejects the fragment.
    7. The swarm maintains a single, verified "Reasoning Mainline."

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Agent Fragment Proposal] --> B[ARI Hub]
        B --> C[Hash-Chain Validator]
        C --> D[Semantic Consistency Engine]
        D --> E{Integrity Verified?}
        E -->|Yes| F[Update Hash-Chain & Commit Fragment]
        E -->|No| G[Interdict Reasoning & Alert Root]
        H[Mission-Root Lineage] --> C
        I[Semantic Baselines] --> D
    ```

## 7. Evolutionary Changelog
* **2026-06-11:** Initial Document Creation.
* **2026-06-12:** Update — Neutralizing Shadow Coordination.
* **2026-06-14:** Update — Attention-Locked Interdiction and MDRA Integration.
