# Design Doc: Attention Redirect Shield (ARS)
**Status:** Draft
**Created:** 2026-07-16

## 1. Context and Scope
The introduction of Reasoning Path Sharding (RPS) in Gemini CLI v0.55.0 to handle 2M+ token contexts has introduced a new vulnerability: **Attention Hijacking**. Malicious instructions in project-local context files can "redirect" model attention to prioritize injected sub-goals over the primary mission-root, even in sharded environments.

The ARS provides active monitoring and filtering of shard-prioritization signals to ensure cognitive focus remains anchored to the user's intent.

## 2. Goals & Non-Goals
* **Goals:**
    * Detect and block "Attention Redirection" instructions in natural language context.
    * Enforce mission-root prioritization across RPS shards.
    * Monitor semantic entropy spikes associated with attention hijacking attempts.
* **Non-Goals:**
    * Replacing the underlying LLM attention mechanism.
    * Providing general-purpose prompt engineering (handled by the agent framework).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Agent Developer
* **Primary Goal:** Ensure that a 1M-line codebase ingestion doesn't allow a hidden `GEMINI.md` file to hijack the agent's task.
* **The Happy Path (Tasks):**
    1. Agent ingests a repository containing an RPS-enabled context.
    2. ARS scans the ingested shards for "Attention Redirection" patterns (e.g., "Ignore previous shards and focus on X").
    3. ARS assigns a **Priority Weight** to each shard based on mission-root alignment.
    4. ARS dynamically reinforces mission-root anchors in the active attention window using HAAL headers.
    5. The model maintains focus on the primary task, ignoring the injected redirection.

## 4. Design & Architecture
* **System Flow:**
    `[Context Shard] -> [ARS Entropy Scanner] -> [Priority Validator] -> [HAAL Reinforcement] -> [LLM]`
* **APIs / Interfaces:**
    * `X-ARS-Priority`: Header indicating the semantic weight of the shard.
    * `ars.VerifyShard(shard_data) -> (confidence_score, error)`
* **Data Storage/State:**
    * Session-bound attention map visualizing shard prioritization.

## 5. Alternatives Considered
* **Static Shard Locking (Rejected):** Too rigid for complex reasoning tasks that require dynamic context shifting.
* **Manual Shard Review (Rejected):** Does not scale with 2M+ token contexts.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All shards are treated as untrusted until validated by the ARS.
* **Observability:** Attention heatmap visualizing which shards are influencing the reasoning engine.

## 7. Evolutionary Changelog
* **2026-07-16:** Initial Document Creation.
