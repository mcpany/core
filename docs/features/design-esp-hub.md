# Design Doc: Epistemic Shard Pinning (ESP) Hub
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
Modern LLMs are prone to "Summarization Erasure," where critical reasoning fragments are lost when context windows are compressed to save tokens. Specifically, the "confidence" or "epistemic weight" of a fragment is often discarded, leading to agents that act on low-certainty information as if it were fact.

The ESP Hub allows agents to "pin" high-certainty shards with hardware-attested metadata. These shards are protected from the summarization engine, ensuring that the cognitive foundation of the mission remains permanent.

## 2. Goals & Non-Goals
* **Goals:**
    * Prevent eviction of high-epistemic-weight fragments during summarization.
    * Standardize confidence scoring across disparate agent frameworks.
    * Implement "Cognitive Anchors" that persist across 1M+ token cycles.
* **Non-Goals:**
    * Automatically determining shard confidence (must be signaled by the agent).
    * Infinite context storage.

## 3. Critical User Journey (CUJ)
* **User Persona:** Long-running Research Agent Orchestrator
* **Primary Goal:** Maintain the "Why" behind a 3-day old data verification step during a final report generation.
* **The Happy Path (Tasks):**
    1. Agent verifies a data point and generates a reasoning fragment.
    2. Agent signals a "Confidence: 0.99" score to the ESP Hub.
    3. ESP Hub generates a hardware-attested pin for that fragment.
    4. The context window hits 128k; the summarization engine compacts the history.
    5. The ESP Hub interdicts and preserves the pinned fragment in its raw form.

## 4. Design & Architecture
* **System Flow:**
    `[Agent Output] -> [Epistemic Scoring Middleware] -> [ESP Hub (Pinning)] -> [Context Summarizer (Exclusion)]`
* **APIs / Interfaces:**
    * `PUT /v1/esp/pin`: Attach a confidence score and hardware-pin to a shard ID.
    * `GET /v1/esp/manifest`: List all currently pinned anchors for a session.
* **Data Storage/State:**
    * Pinned shards are stored in the "Cognitive Sidecar" (SQLite) with cryptographic links to the mission root.

## 5. Alternatives Considered
* **System Prompt Pinning:** Rejected because it pollutes the attention window with static text, losing granularity.
* **Vector Database RAG:** Rejected because RAG is non-deterministic; ESP provides a deterministic guarantee of context presence.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Pins must be signed by the Mission Root to prevent "Shadow Pinning" by rogue subagents.
* **Observability:** UI dashboard shows a "Cognitive Heatmap" of pinned vs. summarized fragments.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
