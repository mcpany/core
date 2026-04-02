# Design Doc: GC-Immune Reasoning Anchors
**Status:** Draft
**Created:** 2026-04-02

## 1. Context and Scope
As AI models transition to 1M+ token context windows (e.g., Claude 4.5, Gemini 2.0 Pro), they have implemented aggressive "Context-Window Garbage Collection" (CWGC) to maintain reasoning performance and reduce KV-cache costs. While efficient, CWGC has introduced a critical stability risk: **Instruction Eviction**.

Users report that core behavioral guardrails and "Mission-Root" intents are being silently pruned from the attention window during long-running sessions, leaving agents susceptible to reasoning drift and instruction injection from subagents. MCP Any must provide infrastructure-level "GC-Immune" pinning to ensure that critical mission anchors remain permanent in the attention window for the duration of the session.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide a mechanism to mark specific context fragments as "GC-Immune."
    * Integrate with model-specific KV-cache priority headers (e.g., `x-gemini-cache-priority`, `x-claude-attention-pin`).
    * Enforce hardware-attested re-verification if an immune fragment is modified.
    * Neutralize "GC Fragility" in deep agent swarms.
* **Non-Goals:**
    * Preventing all context pruning (only protecting marked anchors).
    * Modifying model-side garbage collection logic (leveraging existing APIs).

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise Swarm Architect
* **Primary Goal:** Ensure that "Security Policy X" is never evicted from the attention window during a 48-hour autonomous refactoring mission.
* **The Happy Path (Tasks):**
    1. The Architect defines the "Security Policy" as a GC-Immune Reasoning Anchor in MCP Any.
    2. The Refactoring Swarm initializes and begins generating 100k+ tokens of reasoning traces per hour.
    3. The model's CWGC triggers every 10 minutes to prune low-utility logs.
    4. MCP Any's ALRA middleware ensures the "Security Policy" fragment carries the highest possible retention priority.
    5. At hour 47, the model still adheres to the "Security Policy" despite the vast volume of intervening subagent data.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Mission Root Anchor] --> B[ALRA/GC-Immune Middleware]
        B --> C{Trust Check}
        C -->|Verified| D[Apply Retention Headers]
        D --> E[LLM Attention Layer]
        F[Dynamic Subagent Data] --> G[Context Window]
        G --> H[CWGC Engine]
        H -->|Prune| F
        H -->|Retain| A
    ```
* **APIs / Interfaces:**
    * `POST /v1/anchors/pin`: Marks a text fragment as GC-Immune with a specific priority tier.
    * `GET /v1/anchors/status`: Returns a list of active pinned anchors and their attention scores.
* **Data Storage/State:**
    * Immune anchors are stored in the `anchors` table of the SQLite Blackboard, linked to the hardware-attested mission session.

## 5. Alternatives Considered
* **Periodic Re-Injection**: Rejected because it drastically increases token costs and can cause "Context Confusion" if multiple versions of the same instruction exist in the window.
* **Instruction Summarization**: Rejected because critical security nuances are often lost during summarization.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Only fragments signed by the hardware-attested mission root can be marked as GC-Immune to prevent subagents from "squatting" in the attention window with malicious instructions.
* **Observability:** The "GC-Immune Anchor Visualizer" in the UI highlights pinned fragments and their current attention utilization.

## 7. Evolutionary Changelog
* **2026-04-02:** Initial Document Creation.
