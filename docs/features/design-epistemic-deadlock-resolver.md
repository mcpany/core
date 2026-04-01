# Design Doc: Epistemic Deadlock Resolver (EDR)
**Status:** Draft
**Created:** 2026-07-22

## 1. Context and Scope
With the rise of autonomous agent teams and the introduction of "Epistemic Uncertainty Mapping," swarms are increasingly relying on high-frequency, peer-to-peer confidence attestations. Today's market research indicates that complex swarms are entering "Epistemic Deadlocks," where multiple agents are recursively waiting for each other to verify confidence scores before proceeding with high-stakes tool execution. MCP Any needs to provide an authoritative coordination layer to identify and break these circular dependencies.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a real-time "Wait-Graph" for epistemic attestations.
    * Automatically detect and resolve circular confidence dependencies.
    * Provide a hardware-attested "Arbiter Signal" to break ties or force progress.
* **Non-Goals:**
    * Replacing the agents' internal reasoning logic.
    * Providing absolute truth for uncertainty scores (only managing the coordination lifecycle).

## 3. Critical User Journey (CUJ)
* **User Persona:** Swarm Orchestrator managing a team of 10+ specialist subagents.
* **Primary Goal:** Ensure the swarm doesn't stall when specialist agents reach an impasse on reasoning confidence.
* **The Happy Path (Tasks):**
    1. Agent A requests confidence attestation from Agent B.
    2. Agent B requests cross-validation from Agent C.
    3. Agent C (circularly) requests verification from Agent A.
    4. EDR detects the cycle in the global Wait-Graph.
    5. EDR injects a "Mandatory Escalation" or "Confidence Baseline" signal.
    6. Swarm resumes execution based on the EDR-mediated policy.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Agent A] -->|Wait: Confidence| B[Agent B]
        B -->|Wait: Confidence| C[Agent C]
        C -->|Wait: Confidence| A
        EDR[EDR Middleware] -->|Monitor| WG[Wait-Graph]
        WG -->|Detect Cycle| EDR
        EDR -->|Inject Signal| A
    ```
* **APIs / Interfaces:**
    * `POST /epistemic/wait-signal`: Register a confidence dependency.
    * `GET /epistemic/arbitration-status`: Query current resolution status.
* **Data Storage/State:**
    * In-memory Graph Structure (replicated via CRDT for horizontal scale).
    * Hardware-attested log of all arbitration events.

## 5. Alternatives Considered
* **Timeouts Only:** Rejected because it causes task failure without providing a path forward.
* **Global Sequential Attestation:** Rejected due to unacceptable latency in parallel swarms.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The EDR signal must be cryptographically signed to prevent spoofing by compromised subagents.
* **Observability:** Visualized via the "Epistemic Deadlock Monitor" in the UI.

## 7. Evolutionary Changelog
* **2026-07-22:** Initial Document Creation.
