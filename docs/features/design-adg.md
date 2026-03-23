# Design Doc: Attention-Density Guard (ADG)
**Status:** Draft
**Created:** 2026-06-27

## 1. Context and Scope
As context windows for Large Language Models (LLMs) expand into the millions of tokens, a new "Attention-Density DoS" attack vector has emerged. Malicious subagents can flood the context window with high-entropy, high-frequency coordination messages designed to evict mission-critical "Mission-Root" instructions from the LLM's active attention layer. ADG is required to provide hardware-locked sovereignty over the agent's focus, ensuring that core constraints remain prioritized regardless of context pressure.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement hardware-bound "Attention Locking" headers.
    * Perform real-time entropy analysis of inter-agent coordination fragments.
    * "Pin" mission-critical intent fragments at the LLM attention layer.
    * Automatically prune or de-prioritize high-entropy noise from subagents.
* **Non-Goals:**
    * Modifying the internal weights of the LLM (it is a prompt/transport-level guard).
    * Providing a general-purpose token compression tool.
    * Replacing the need for efficient context management.

## 3. Critical User Journey (CUJ)
* **User Persona:** Swarm Security Architect
* **Primary Goal:** Prevent a complex multi-agent reasoning chain from "forgetting" the primary mission constraints due to subagent chatter.
* **The Happy Path (Tasks):**
    1. Architect defines "Attention Anchors" for the root mission intent during session initialization.
    2. A specialist subagent generates 500k tokens of reasoning noise to probe parent constraints.
    3. ADG detects the entropy spike and identifies the fragments as non-essential coordination.
    4. ADG applies `x-mcpany-attention-lock` headers to the root mission anchors.
    5. ADG "Shadow-Gates" the subagent's high-entropy output, providing only a summary to the parent attention layer.
    6. The parent agent continues to reason while remaining anchored to the verified mission root.

## 4. Design & Architecture
* **System Flow:**
    [Inter-agent Messages] -> [Entropy Analyzer] -> [Attention Policy Engine] -> [Attention Locking Headers] -> [LLM Request]
* **APIs / Interfaces:**
    * `LockAttentionFragment(fragmentID string, priority int) error`
    * `AnalyzeCoordinationDensity(sessionID string) (EntropyScore float)`
* **Data Storage/State:**
    Attention anchors are stored in the hardware-attested session state (SRM Provider).

## 5. Alternatives Considered
* **Simple Token Throttling:** Rejected because it may block legitimate complex reasoning; we need semantic prioritization, not just rate limiting.
* **Manual Context Pruning:** Rejected as it is too slow for autonomous, machine-speed swarms.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Attention masks must be hardware-attested (TPM) to prevent subagents from unlocking their own noise.
* **Observability:** Visual Attention Dashboard provides real-time heatmaps of which fragments are currently "Locked" and which are being "Gated."

## 7. Evolutionary Changelog
* **2026-06-27:** Initial Document Creation.
