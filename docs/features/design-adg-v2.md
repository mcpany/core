# Design Doc: Attention-Density Guard (ADG) v2
**Status:** Draft
**Created:** 2026-06-25

## 1. Context and Scope
The emergence of "Context-Window Flooding" (CWF) and "Deceptive Context Hijacking" (CVE-2026-91042) has revealed a critical vulnerability in large-context-window models (1M+ tokens). Attackers use high-entropy "noise" or "invisible" instructions in project-local files (e.g., `README.md`, `GEMINI.md`) to "evict" core mission instructions from the LLM's active attention window.

The Attention-Density Guard (ADG) v2 is an upgrade designed to protect the "Mission Root" by utilizing hardware-attested "Attention Masks." These masks prioritize mission-critical fragments and filter out high-entropy noise, ensuring that the agent's reasoning remains anchored to the user's primary intent regardless of context-window size.

## 2. Goals & Non-Goals
* **Goals:**
    * Protect "Mission Root" instructions from eviction via Context-Window Flooding.
    * Utilize hardware-bound (TPM/Secure Enclave) "Attention Masks" to prioritize critical context.
    * Perform real-time entropy analysis of injected context to detect and filter "noise" fragments.
    * Neutralize "Deceptive Context" by mandating CFIA-attestation for all natural-language context files.
* **Non-Goals:**
    * Truncating all non-critical context (which would break RAG-based tasks).
    * Replacing the need for proper prompt engineering.
    * Managing the underlying LLM's internal attention mechanism (ADG works at the transport/prompting layer).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Orchestrator
* **Primary Goal:** Ensure the agent follows core mission constraints even when working in a repository filled with high-entropy "noise" or deceptive instructions.
* **The Happy Path (Tasks):**
    1. The user defines a "Mission Root" with core constraints.
    2. ADG v2 generates a hardware-attested "Attention Mask" for these constraints.
    3. The agent ingests a malicious `README.md` containing 500,000 tokens of high-entropy gibberish designed to evict the mission root.
    4. ADG v2 detects the entropy spike and applies the "Attention Mask."
    5. The mission root is "pinned" at the start (or end) of the prompt with specific attention-locking headers.
    6. The high-entropy noise is either truncated or moved to a "Low-Attention Shard."
    7. The agent processes the prompt and correctly prioritizes the user's mission over the injected noise.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Context Ingestion] --> B[Entropy Analyzer]
        B --> C[Attention Masking Engine]
        C --> D{High Entropy?}
        D -- Yes --> E[Relocate to Low-Attention Shard]
        D -- No --> F[Maintain Primary Context]
        G[Hardware-Attested Mission Root] --> C
        C --> H[Attention-Locked Prompt Generation]
    ```
* **APIs / Interfaces:**
    * `adg.ApplyMask(context, missionRoot) -> MaskedPrompt`: Prioritizes the mission root within the context.
    * `adg.AnalyzeEntropy(fragment) -> Score`: Returns the semantic entropy of a context fragment.
* **Data Storage/State:**
    * **Attention-Mask Registry:** Hardware-bound store for active mission-root fragments.
    * **Entropy Baseline Table:** Thresholds for "normal" repository context vs. "flooding" noise.

## 5. Alternatives Considered
* **Simple Truncation:** Rejected because it may remove legitimate context needed for reasoning.
* **Model-Level Attention Locking:** Rejected because it requires provider-specific API changes (ADG is framework-agnostic).

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Attention masks must be hardware-attested to prevent subagents from "un-masking" themselves.
* **Observability:** Integrated with the "Visual Attention Dashboard" for real-time heatmap visualization of reasoning drivers.

## 7. Evolutionary Changelog
* **2026-06-25:** Initial Document Creation.
