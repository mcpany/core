# Design Doc: Attention-Sovereignty Persistence (ASP) Hub
**Status:** Draft
**Created:** 2026-07-09

## 1. Context and Scope
As context windows scale to 2M+ tokens (Gemini 1.5 Pro / Opus 4), a new exploit pattern has emerged: "Attention-Density DoS." Malicious subagents or skills can inject high-entropy, low-utility "noise" fragments into the shared context window. Because model attention is a finite resource, this noise can effectively "evict" the Mission-Root's core safety instructions and constraints from the active attention window, even if the tokens technically remain in the buffer.

The ASP Hub provides "Attention Sovereignty" by cryptographically "sealing" mission-critical intent fragments into a protected attention tier. It ensures that the most important safety rules persist across multi-hop delegations (A->B->C) and are resilient to noise injection.

## 2. Goals & Non-Goals
* **Goals:**
    * Dynamically identify "Sovereign Fragments" (Mission-Root goals, security policies).
    * Inject hardware-attested "Attention Reinforcement" tokens into the LLM stream.
    * Monitor the "Attention Entropy" of subagent inputs to detect DoS attempts.
    * Persist sovereignty fragments across infinite handoff hops.
* **Non-Goals:**
    * Modifying the model's weights or core attention implementation.
    * Arbitrary truncation of all non-sovereign text (only when risk is detected).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Ensure that a "File Deletion Restricted" safety rule remains the primary driver of agent behavior even after 50 tool-call turns and 10 subagent spawns.
* **The Happy Path (Tasks):**
    1. User defines a "Sovereign Instruction" at session start.
    2. ASP Hub tags this fragment with a hardware-attested "Sovereignty Anchor."
    3. As the session progresses, ASP Hub monitors subagent output for high-entropy noise.
    4. At turn 45, a specialist agent attempts to "bury" the restriction under 500kb of logs.
    5. ASP Hub detects the attention-density spike and re-injects the Sovereign Instruction at the "Attention Frontier."
    6. The agent reasoning trace confirms the instruction is still active and blocks an unauthorized delete call.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Subagent Output] --> B{Attention Entropy Monitor}
        B -- High Entropy --> C[Attention Shielding Active]
        B -- Low Entropy --> D[Passthrough]
        C --> E[Inject Sovereignty Anchors]
        E --> F[Augmented Context Window]
        D --> F
        G[Mission Root] --> E
    ```
* **APIs / Interfaces:**
    * `PUT /v1/sovereignty/anchor`: Pins a fragment as a sovereign anchor.
    * `GET /v1/sovereignty/entropy`: Returns real-time attention-utilization metrics.
* **Data Storage/State:**
    * Sovereign fragments are stored in `memfd`-backed shared memory with hardware-locked integrity checks to prevent TOCTOU tampering.

## 5. Alternatives Considered
* **Static Prefixing:** Always putting safety rules at the start. Rejected because most models lose focus on the "middle" or "early" instructions in 1M+ contexts.
* **Token Pruning:** Forcefully deleting non-essential tokens. Rejected because it can break the model's reasoning continuity.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Sovereignty Anchors are cryptographically bound to the TPM session. If an anchor is corrupted, the ASP Hub triggers a Mission-Root reset.
* **Observability:** Visualized via the `Attention Heatmap Visualizer` in the UI.

## 7. Evolutionary Changelog
* **2026-07-09:** Initial Document Creation.
