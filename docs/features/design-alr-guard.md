# Design Doc: Attention-Locked Reasoning (ALR) Guard
**Status:** Draft
**Created:** 2026-06-25

## 1. Context and Scope
With the rise of "Attention-Eviction" attacks, where adversaries use high-entropy "Noise Injections" to force mission-critical instructions out of an LLM's active context window, MCP Any needs a mechanism to ensure "Attention Sovereignty."

The Attention-Locked Reasoning (ALR) Guard is a core security middleware that utilizes hardware-attested attention-locking headers to "pin" the most sensitive mission-root intent fragments at the LLM's attention layer.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Protect mission-root intents from context-window eviction by adversarial noise.
    *   Provide hardware-attested proof of "Attention Persistence."
    *   Support dynamic, attention-weighted prioritization of context shards.
*   **Non-Goals:**
    *   Solving generic context-window size limitations.
    *   Providing absolute immunity to all forms of reasoning hijacking.

## 3. Critical User Journey (CUJ)
*   **User Persona:** Enterprise Swarm Administrator
*   **Primary Goal:** Ensure that a deep swarm of 50+ subagents cannot accidentally "flush" the primary security policy while processing massive datasets.
*   **The Happy Path (Tasks):**
    1.  The administrator defines the "Mission Root" intent in the HAMM manifest.
    2.  MCP Any signs the intent and attaches the `x-mcp-attention-lock: high` header.
    3.  During subagent execution, the ALR Guard monitors the attention-utilization of the LLM.
    4.  If noise-entropy exceeds a safety threshold, the ALR Guard prunes low-priority subagent context while keeping the "Attention-Locked" intent pinned.
    5.  The primary security policy remains the dominant driver of the agent's tool-call reasoning.

## 4. Design & Architecture
*   **System Flow:**
    `[Subagent Reasoner] -> [ALR Guard Middleware] -> [HART-Attested Provider] -> [LLM Gateway]`
*   **APIs / Interfaces:**
    *   `mcp_any.v1.AttentionService`: Manages attention-locked shards.
    *   `x-mcp-attention-lock`: A hardware-attested header for transport-layer attention pinning.
*   **Data Storage/State:**
    *   The ALR Guard stores mission-root "Attention Anchors" in a TPM-locked shard of the Shared KV Store (Blackboard).

## 5. Alternatives Considered
*   **Generic Context Summarization**: Rejected because generic summaries often lose the strict "Semantic Sovereignty" required for security-critical intents.
*   **Static Attention Masking**: Rejected because it cannot adapt to the real-time reasoning-entropy of a dynamic swarm.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** The ALR Guard utilizes hardware-attested signatures to prevent subagents from "Self-Unpinning" mission-critical intents.
*   **Observability:** Integrated with the Visual Attention Dashboard to provide real-time heatmaps of pinned vs. evicted context fragments.

## 7. Evolutionary Changelog
*   **2026-06-25:** Initial Document Creation.
