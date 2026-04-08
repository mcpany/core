# Design Doc: Mission-Root Defragmenter (MRD)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As AI agent swarms scale to long-running missions with 1M+ token context windows, they suffer from "Mission Root Erosion" (the Ship of Theseus effect). Recursive subagent rotations, summarization events, and "attention drift" cause the primary user intent (the Mission Root) to be semantically diluted or evicted from the model's active attention window.

The Mission-Root Defragmenter (MRD) is an active service in MCP Any that periodically restores the semantic purity of the reasoning loop. It forcefully prunes high-entropy "semantic noise" and re-injects hardware-attested mission anchors, ensuring that core behavioral guardrails remain permanent and prioritized.

## 2. Goals & Non-Goals
* **Goals:**
    * Perform real-time semantic density analysis of the agent's attention window.
    * Identify and prune "Ghost Fragments" that contribute to intent drift.
    * Dynamically re-inject fresh, hardware-attested mission-root fragments into the reasoning loop.
    * Synchronize defragmentation events across parallel teammate shards.
* **Non-Goals:**
    * Replacing general-purpose context summarization.
    * Modifying the underlying model's attention mechanism (it works at the prompt/context injection layer).

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise Swarm Architect
* **Primary Goal:** Maintain strict security guardrails in a multi-day autonomous research mission.
* **The Happy Path (Tasks):**
    1. User initiates a mission with a signed "Mission Root" manifest.
    2. Swarm operates for 6 hours, spawning 15 specialist subagents.
    3. MRD detects that mission-root semantic density has fallen below the 15% threshold due to noise from specialist reasoning traces.
    4. MRD triggers a "Defragmentation Event."
    5. MRD prunes non-essential specialist monologues and prepends the original hardware-attested manifest to the active context.
    6. Swarm resumes operation with restored behavioral anchors.

## 4. Design & Architecture
* **System Flow:**
    `[Context Engine] -> [MRD Density Analyzer] -> [Semantic Scrubber] -> [Manifest Re-Injector] -> [Model Inference]`
* **APIs / Interfaces:**
    * `mrd.AnalyzeDensity(context_fragments)`: Returns entropy/utility scores.
    * `mrd.TriggerDefrag(session_id)`: Forcefully realigns the session context.
* **Data Storage/State:** MRD utilizes the Shared KV Store (Blackboard) to track the "Semantic Lineage" of context fragments.

## 5. Alternatives Considered
* **Static Pinning (ALRA)**: Rejected as the sole solution because models can still "ignore" pinned content if the surrounding noise entropy is too high. Active defragmentation (pruning) is required.
* **Frequent Re-Summarization**: Rejected because aggressive summarization often *causes* the intent drift by flattening critical semantic nuances.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Re-injected manifests must match the original hardware signature (TPM-signed) to prevent "Manifest Spoofing."
* **Observability:** Defragmentation events and "pruned fragment" logs are surfaced in the Visual Attention Dashboard.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
