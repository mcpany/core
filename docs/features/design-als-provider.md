# Design Doc: Attention-Locked Sovereignty (ALS) Provider
**Status:** Draft
**Created:** 2026-06-18

## 1. Context and Scope
In horizontal teammate coordination, the "Attention Window" of the LLM has become a contested resource. The emergence of "Context-Window Flooding" (CWF) attacks—where malicious subagents inject high-entropy, noisy reasoning fragments—can lead to the eviction of critical mission-root instructions from the agent's attention loop. This "Attention Hijacking" results in intent drift and loss of control.

The Attention-Locked Sovereignty (ALS) Provider addresses this by utilizing hardware-bound attention-locking headers. It ensures that the "Mission Root" intent is pinned at the highest priority tier of the agent's attention mechanism, rendering it immune to eviction by subagent noise.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a priority-based attention governance service.
    * Mandate "ALS-locking" for all mission-root intents.
    * Provide hardware-attested "Attention Manifests" to verify instruction priority.
    * Integrate with the ARI Hub to detect and block "Attention Hijacking" attempts.
* **Non-Goals:**
    * Expanding the physical context window (limited by the model provider).
    * Re-writing subagent messages (they are gated/pruned, not modified).
    * Managing token budgets (handled by the Reasoning-Budget Firewall).

## 3. Critical User Journey (CUJ)
* **User Persona:** Swarm Security Architect
* **Primary Goal:** Protect the core mission objectives from being "drowned out" by a specialist agent that has entered a high-entropy reasoning loop.
* **The Happy Path (Tasks):**
    1. Parent Agent initializes a mission with ALS-locking enabled.
    2. The ALS Provider wraps the core intent in hardware-bound priority headers (`X-ALS-Tier: Root`).
    3. A specialist subagent generates a massive, high-entropy reasoning trace (either accidentally or maliciously).
    4. The LLM's attention mechanism processes the subagent noise but keeps the `Root` tier instructions pinned.
    5. The ALS Provider monitors the attention-utilization score.
    6. If the mission root is threatened, the ALS Provider signals the Dynamic Attention Gating (DAG) middleware to prune the subagent's low-priority noise.
    7. Mission sovereignty is maintained.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Mission Intent] --> B[ALS Provider]
        B --> C[ALS-Locking: Tier 0]
        D[Subagent Fragments] --> E[ALS Provider]
        E --> F[Assign Tier: N+1]
        C --> G[LLM Attention Mechanism]
        F --> G
        G --> H{Attention Utilization High?}
        H -- Yes --> I[Trigger DAG Pruning of Tier > 0]
        H -- No --> J[Continue Processing]
    ```
* **APIs / Interfaces:**
    * `als.LockIntent(intent, tier) -> LockedFragment`: Applies hardware-bound priority headers.
    * `als.GetAttentionManifest(sessionID) -> Manifest`: Returns the current priority tiers.
* **Data Storage/State:**
    * **Attention Manifest Registry:** Ephemeral state tracking the priority tier assigned to every fragment in the current context window.

## 5. Alternatives Considered
* **Semantic Summarization:** Rejected because summarization can lose critical nuances of the mission root.
* **Fixed Token Reservation:** Rejected as it lacks the dynamic priority needed for horizontal meshes.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** ALS headers are hardware-bound and session-specific. Any attempt to spoof a `Tier 0` header by a subagent triggers an immediate session termination.
* **Observability:** Attention-utilization and pruning events are visualized in the "Context Attention Monitor."

## 7. Evolutionary Changelog
* **2026-06-18:** Initial Document Creation.
