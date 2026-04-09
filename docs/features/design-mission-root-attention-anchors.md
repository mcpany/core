# Design Doc: Mission-Root Attention Anchors (MRAA)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
With the adoption of 1M+ token context windows (e.g., Gemini 1.5 Pro), AI agents are increasingly susceptible to "Instruction Eviction". As high-volume tool data or reasoning traces fill the attention window, the primary mission-root guardrails (system instructions) can be de-prioritized or completely evicted by the model's attention mechanism. MRAA ensures that critical instructions remain "permanently anchored" and immune to garbage collection.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement mandatory pinning for specific mission-root fragments.
    * Provide "GC-Immune" status for instructions that define security boundaries.
    * Support hardware-attested re-injection of anchors if window saturation is detected.
    * Neutralize "Attention Hijacking" where subagent noise evicts parent guardrails.
* **Non-Goals:**
    * Managing the entire context window (MRAA focuses only on the "Anchors").
    * Replacing model-native attention mechanisms.

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise Security Admin
* **Primary Goal:** Ensure that an agent analyzing 10,000 log files never "forgets" the instruction to never exfiltrate data.
* **The Happy Path (Tasks):**
    1. Admin defines a "Sovereign Instruction Set" (SIS) for the mission.
    2. MCP Any marks these fragments with the `x-mcpany-anchor-priority: critical` header.
    3. The agent processes a massive dataset, exceeding its "natural" attention peak.
    4. MRAA middleware detects attention drift or fragment eviction signals from the model adapter.
    5. MCP Any programmatically re-synchronizes the SIS at the "Freshness Boundary" of the window.
    6. Behavioral integrity is maintained throughout the 1M+ token reasoning cycle.

## 4. Design & Architecture
* **System Flow:**
    `Mission Root` -> `MRAA Manager` -> `Anchor Pinning` -> `Attention Monitor` -> `Re-injection Loop`
* **APIs / Interfaces:**
    * `AnchorRegistry`: `RegisterAnchor(fragment Fragment) error`
    * `AttentionObserver`: `OnEviction(anchorID string) callback`
* **Data Storage/State:**
    * SIS fragments stored in TPM-protected kernel memory.

## 5. Alternatives Considered
* **Frequent Re-prompting**: Rejected as it consumes excessive tokens and disrupts reasoning coherence.
* **Context Truncation**: Rejected as it leads to information loss in deep RAG/Analysis tasks.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: Only the mission-root can define or modify SIS anchors.
* **Observability**: Real-time "Attention Map" in the dashboard shows the prominence of SIS anchors.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
