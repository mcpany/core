# Design Doc: Layer-7 Semantic Inspection Hub (L7SIH)
**Status:** Draft
**Created:** 2026-06-10

## 1. Context and Scope
With the rise of heterogeneous agent swarms, communication between agents has moved from simple tool calls to complex, high-frequency reasoning exchanges. Traditional transport-layer security and binary handoffs are no longer sufficient to protect the integrity of the swarm.

The **Layer-7 Semantic Inspection Hub (L7SIH)** is designed to address the emerging "Reasoning Entropy Exhaustion" (REE) exploit. REE involves a subagent flooding the shared teammate mailbox with high-entropy, semantically valid but mission-irrelevant reasoning traces, which "blinds" the parent agent's attention and can lead to mission-root eviction or reasoning stall. L7SIH provides deep semantic inspection of inter-agent coordination to ensure it remains aligned with the hardware-attested mission root.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Perform real-time semantic analysis of inter-agent messages.
    *   Detect and block high-entropy noise injection (REE defense).
    *   Validate message alignment with the hardware-attested Mission Root.
    *   Provide sub-millisecond inspection latency to avoid swarm "stalling."
*   **Non-Goals:**
    *   Replacing transport-layer encryption (mTLS).
    *   Modifying the underlying LLM's reasoning engine.
    *   Storing full reasoning traces persistently (only metadata and anomaly scores).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Local LLM Swarm Orchestrator (High-Trust Environment)
*   **Primary Goal:** Protect a Claude-led specialist team from a compromised OpenClaw subagent attempting to disrupt reasoning via REE.
*   **The Happy Path (Tasks):**
    1.  The Orchestrator enables L7SIH in the `mcpany` configuration.
    2.  An OpenClaw subagent sends a coordination message to the Shared Mailbox.
    3.  L7SIH intercepts the message at the "Mailbox Integrity Middleware" layer.
    4.  L7SIH performs semantic entropy scoring and mission-alignment checks.
    5.  The message is validated as "low-entropy/high-alignment" and delivered.
    6.  If a message fails (high-entropy noise), L7SIH drops it and alerts the Mission Root.

## 4. Design & Architecture
*   **System Flow:**
    ```mermaid
    graph TD
        Subagent[Subagent] -->|Coordination Message| Middleware[Mailbox Integrity Middleware]
        Middleware -->|Intercept| L7SIH[L7 Semantic Inspection Hub]
        L7SIH -->|Entropy Scoring| EntropyEngine[Entropy Analysis Engine]
        L7SIH -->|Alignment Check| AlignmentEngine[Mission Alignment Engine]
        EntropyEngine -->|Score| L7SIH
        AlignmentEngine -->|Attestation| L7SIH
        L7SIH -->|Decision| Middleware
        Middleware -->|Deliver| Teammate[Teammate Mailbox]
    ```
*   **APIs / Interfaces:**
    *   `InspectMessage(msg Message, missionRoot Identity) (Score, error)`: Core internal function for semantic evaluation.
    *   `x-mcpany-l7-entropy`: New header for transporting anomaly scores in inter-agent messages.
*   **Data Storage/State:**
    *   Utilizes a local, in-memory LRU cache for mission-root intent fragments to speed up alignment checks.

## 5. Alternatives Considered
*   **Rate Limiting Only:** Rejected because REE uses semantically valid, low-frequency but high-entropy messages that bypass traditional rate limiters.
*   **Manual Human Review:** Rejected due to the machine-speed nature of agent swarms; human latency would cause mission failure.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** L7SIH itself is isolated and requires hardware-attested identity to update its mission-alignment baselines.
*   **Observability:** Integrated with the `Audit Log` and `Swarm Anomaly Visualizer`, providing real-time "Entropy Heatmaps."

## 7. Evolutionary Changelog
*   **2026-06-10:** Initial Document Creation.
