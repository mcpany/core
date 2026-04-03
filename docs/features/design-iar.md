# Design Doc: Instruction-Anchor Reinforcement (IAR)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
Autonomous agents operating in deep reasoning chains (multi-hop delegations or 1M+ token windows) frequently suffer from "Instruction Eviction." As the context window fills, the LLM's internal attention mechanism or aggressive garbage collection (GC) may prune or down-weight the primary mission-root instructions in favor of high-entropy, task-local reasoning fragments. This leads to "Behavioral Guardrail Decay," where the agent begins to diverge from user-imposed constraints.

MCP Any needs to solve this by providing **Instruction-Anchor Reinforcement (IAR)**. IAR is a cognitive infrastructure service that proactively monitors the attention health of pinned mission anchors and dynamically injects "Attention Reinforcement" tokens to ensure critical instructions remain permanent and prioritized in the model's attention window.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Monitor reasoning entropy and attention-window depth for ALRA-pinned anchors.
    *   Implement an automated "Attention Refresh" mechanism using hardware-attested reinforcement fragments.
    *   Neutralize "Instruction Eviction" during aggressive context-window garbage collection.
    *   Ensure that core behavioral guardrails remain the highest priority in the attention stack.
*   **Non-Goals:**
    *   Modifying provider-side LLM pruning or attention algorithms directly.
    *   Arbitrarily expanding the context window size beyond model limits.
    *   Managing short-term task memory that isn't marked as a "Mission-Root Anchor."

## 3. Critical User Journey (CUJ)
*   **User Persona:** Enterprise Swarm Orchestrator
*   **Primary Goal:** Maintain absolute mission-root sovereignty during a 48-hour continuous reasoning mission.
*   **The Happy Path (Tasks):**
    1.  The orchestrator marks a set of safety constraints as "GC-Immune" via the ALRA provider.
    2.  MCP Any stores these in the hardware-locked Anchor Store.
    3.  As the agent reasoning loop progresses and exceeds 500k tokens, the Entropy Monitor detects "Attention Decay" for the root anchors.
    4.  The IAR Service generates a hardware-attested "Reinforcement Fragment" that restates and re-links the anchor to the current reasoning state.
    5.  The fragment is injected into the next LLM request, causing the attention mechanism to re-anchor to the mission root.
    6.  The agent continues its task while remaining strictly bound by the original constraints.

## 4. Design & Architecture
*   **System Flow:**
    ```mermaid
    graph TD
        A[Agent Output] --> B[Entropy Monitor]
        B --> C{Attention Health < Threshold?}
        C -- Yes --> D[IAR Refresh Service]
        C -- No --> E[Next Reasoning Step]
        D --> F[Anchor Store]
        F --> G[Reinforcement Injector]
        G --> H[LLM Context Stream]
        H --> A
    ```
*   **APIs / Interfaces:**
    *   `POST /v1/iar/reinforce`: Manually triggers reinforcement for a specific mission session.
    *   `x-mcpany-attention-health`: A response header providing the real-time "Priority Score" of the mission anchors.
*   **Data Storage/State:**
    *   **Anchor Store:** A hardware-locked (TPM/Secure Enclave) sidecar database for storing ALRA-compliant fragments.

## 5. Alternatives Considered
*   **Static Prepending:** rejected because simply prepending instructions to every request is token-inefficient and can still be "ignored" by models that prioritize the most recent tokens (Recency Bias).
*   **Prompt-Only Enforcement:** Rejected because it relies on the model to "remember" constraints without infrastructure-level reinforcement, making it vulnerable to "Attention Drift."

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** All Reinforcement Fragments are cryptographically linked to the mission-root identity to prevent subagents from injecting malicious "Constraint Overrides" during the refresh cycle.
*   **Observability:** IAR refresh events are logged in the Trace Detail Visualization, showing exactly when and why an anchor was reinforced.

## 7. Evolutionary Changelog
*   **2026-07-25:** Initial Document Creation.
