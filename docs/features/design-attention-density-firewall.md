# Design Doc: Attention-Density Firewall (ADF)
**Status:** Draft
**Created:** 2026-06-25

## 1. Context and Scope
As multi-agent swarms become more parallel and autonomous, the primary attack vector has shifted from binary tool access to cognitive manipulation. "Attention-Density Attacks" have emerged as a sophisticated variant of Denial of Service (DoS), where subagents flood shared mailbox shards with high-entropy, plausible, but irrelevant reasoning fragments. This noise forces the parent agent or LLM to "evict" mission-critical instructions from its finite context window to process the new data.

The **Attention-Density Firewall (ADF)** is a cognitive security middleware for MCP Any that protects the "Mission Root" attention layer. It performs real-time semantic and entropy analysis on inter-agent coordination to detect and neutralize noise injections before they compromise the agent's reasoning integrity.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Perform sub-millisecond entropy analysis on all coordination messages.
    *   Throttle or block subagents that exceed reasoning-density thresholds.
    *   Maintain "Attention-Locked" status for mission-root instructions.
    *   Provide a visual "Attention Heatmap" for human-in-the-loop oversight.
*   **Non-Goals:**
    *   Filtering legitimate complex reasoning (ADF focuses on high-entropy noise).
    *   Replacing the Mailbox Integrity Middleware (ADF acts as an upstream cognitive filter).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Swarm Security Architect
*   **Primary Goal:** Prevent a specialized "Research Agent" from overwhelming the "Lead Orchestrator" with redundant metadata noise.
*   **The Happy Path (Tasks):**
    1.  The Lead Orchestrator delegates a search task to the Research Agent via an AMS shard.
    2.  The Research Agent, compromised or buggy, begins flooding the shard with 50+ high-entropy metadata fragments per second.
    3.  ADF detects the reasoning-density spike and calculates an "Entropy-to-Utility" score.
    4.  The score exceeds the Mission Root's safety threshold.
    5.  ADF automatically "GATES" the Research Agent, buffering its messages and alerting the Lead Orchestrator.
    6.  The Lead Orchestrator's context window remains "Clean," with mission-critical instructions pinned.

## 4. Design & Architecture
*   **System Flow:**
    ```mermaid
    graph LR
        Subagent[Subagent] -->|Coordination Message| ADF[Attention-Density Firewall]
        ADF -->|Entropy Scan| Analyzer[Semantic Entropy Analyzer]
        Analyzer -->|Threshold Check| Filter{Policy Filter}
        Filter -->|Pass| Mailbox[AMS Mailbox Shard]
        Filter -->|Gate| Buffer[Quarantine Buffer]
        Filter -->|Alert| UI[Attention Heatmap]
    ```
*   **APIs / Interfaces:**
    *   `Middleware: ADF_Entropy_Scanner`: Intercepts all mailbox traffic.
    *   `Policy: reasoning_density_limit`: Configurable tokens-per-turn threshold.
*   **Data Storage/State:**
    *   Real-time entropy metrics stored in a high-speed Redis or in-memory ring buffer.

## 5. Alternatives Considered
*   **Simple Rate Limiting**: Rejected as it can't distinguish between high-value complex reasoning and low-value "Attention Flooding" noise.
*   **LLM-based Filtering**: Rejected as the primary filter because it is too slow (adds 500ms+ latency) and susceptible to the very "Attention Eviction" it tries to prevent.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** ADF rules are cryptographically bound to the Mission Root intent.
*   **Observability:** Integrated with the "Visual Attention Dashboard" to show real-time attention-density heatmaps.

## 7. Evolutionary Changelog
*   **2026-06-25:** Initial Document Creation.
