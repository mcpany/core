# Design Doc: PAM (Pre-computed Attention Mask) Broker
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
With the scaling of LLM context windows to 2M+ tokens, agents in high-density meshes are suffering from "Attention Leakage," where irrelevant background context or noise fragments distract the model from core mission-root instructions. Gemini CLI v0.60.0 introduced hardware-layer support for Attention Masks. MCP Any needs a PAM Broker to pre-calculate and enforce these masks at the infrastructure level, ensuring subagents only attend to the context fragments authorized by the Mission Root.

## 2. Goals & Non-Goals
* **Goals:**
    * Pre-calculate attention masks based on hardware-attested mission manifests.
    * Intercept subagent tool calls and inject the appropriate `x-gemini-attention-mask` header.
    * Reduce model reasoning effort (ARE) and latency by pruning irrelevant context shards.
    * Prevent subagents from probing unauthorized context fragments through the attention layer.
* **Non-Goals:**
    * Modifying the model's underlying attention mechanism (this is a broker/proxy service).
    * Handling context fragments that are not managed by the MCP Any state layer.

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Ensure a specialized "Database Agent" only attends to the relevant schema fragments and its own task instructions, ignoring 1M tokens of unrelated "Web Search" logs in the shared window.
* **The Happy Path (Tasks):**
    1. The Mission Root defines a Mission Manifest with context shard mappings.
    2. The PAM Broker calculates a bit-mask for authorized shards.
    3. The Database Agent initiates a reasoning turn.
    4. PAM Broker intercepts the request and injects the pre-computed hardware-bound mask.
    5. The LLM processes only the authorized attention heads, reducing ARE.

## 4. Design & Architecture
* **System Flow:**
    `[Agent] -> [PAM Broker Middleware] -> [Attention Mask Generator] -> [Hardware Gateway (TPM/GPU)]`
* **APIs / Interfaces:**
    * `GenerateMask(mission_id, shard_list)`: Internal calculation endpoint.
    * `x-mcp-attention-anchor-id`: Mapping token for shard-to-mask resolution.
* **Data Storage/State:**
    Attention masks are cached in the Shared KV Store (Blackboard) and linked to hardware enclave IDs.

## 5. Alternatives Considered
* **Client-Side Masking:** Rejected because subagents can bypass masks to exfiltrate unauthorized fragments.
* **Dynamic Masking during Inference:** Rejected due to the high latency tax on every token generation.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Masks are cryptographically bound to the mission-root manifest; subagents cannot expand their attention field without re-attestation.
* **Observability:** Integration with the `Attention Anchor Heatmap` for visual verification of model focus.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
