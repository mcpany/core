# Design Doc: Hardware-Attested Attention (HAA) Bridge
**Status:** Draft
**Created:** 2026-06-22

## 1. Context and Scope
The introduction of Gemini CLI v0.42.0 and the emergence of "Attention-Window Flooding" attacks reveal a new critical failure point in multi-agent swarms. Subagents can use high-entropy reasoning noise to "evict" mission-root anchors from the parent agent's attention window, leading to intent drift and security bypasses.

The **Hardware-Attested Attention (HAA) Bridge** acts as an authoritative "Attention Guard" for MCP Any. It mandates hardware-bound (TPM) headers to cryptographically "lock" mission-critical intent fragments at the LLM attention layer, ensuring they remain prioritized throughout the reasoning lifecycle.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Protect mission-root anchors from eviction via high-entropy subagent noise.
    *   Enforce hardware-bound (TPM) attention locking for critical intent fragments.
    *   Provide real-time "Attention-Utilization" scores to the Dynamic Attention Gating (DAG) middleware.
    *   Standardize attention-locking headers across framework-neutral handoffs.
*   **Non-Goals:**
    *   Replacing the LLM's internal attention mechanism.
    *   Managing global context windows (handled by the ContextEngine).
    *   Fixing model-specific context window limitations.

## 3. Critical User Journey (CUJ)
*   **User Persona:** High-Trust Swarm Developer
*   **Primary Goal:** Ensure that a "Write-Deny" mission-root constraint remains effective even when subagents generate thousands of lines of code.
*   **The Happy Path (Tasks):**
    1.  The developer defines a mission-root anchor with the `haa-lock: true` property in the mission manifest.
    2.  MCP Any generates a hardware-attested HAA token via the TPM.
    3.  During LLM invocation, the HAA Bridge injects the `x-mcp-haa-lock` header bound to the token.
    4.  A subagent attempts to "flood" the context window with reasoning noise.
    5.  The LLM's attention layer, respecting the HAA lock, prioritizes the mission-root anchor, preventing "Attention Eviction."

## 4. Design & Architecture
*   **System Flow:**
    ```mermaid
    graph TD
        Agent[Primary Agent] -->|Define Anchor| HAA[HAA Bridge]
        HAA -->|Request Attestation| TPM[Hardware TPM]
        TPM -->|Signed Lock Token| HAA
        HAA -->|Inject Lock Header| LLM[LLM / Inference Engine]

        Subagent -->|Inject Noise| LLM
        LLM -->|Attention Priority| Anchor[Locked Mission Anchor]
        LLM -->|Attention Pruning| Noise[Subagent Noise]
    ```
*   **APIs / Interfaces:**
    *   `POST /attention/lock`: Registers a context fragment for hardware-attested attention locking.
    *   `GET /attention/status`: Returns real-time attention-utilization scores.
    *   `Header: x-mcp-haa-token`: Carries the hardware signature for the attention lock.
*   **Data Storage/State:**
    *   HAA lock states are held in memory-mapped volatile storage for sub-millisecond access.
    *   TPM-bound signatures are cached per-mission session.

## 5. Alternatives Considered
*   **Soft Pinning (System Prompts)**: Rejected as system prompts can still be "shadowed" or diluted by massive subagent context injections.
*   **Context Truncation**: Rejected as it is non-discriminatory and can remove legitimate high-importance subagent state.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** Attention locks are cryptographically bound to the hardware Inode of the mission intent to prevent "Lock Spoofing."
*   **Observability:** Integrated with the "Visual Attention Dashboard" to show real-time heatmap of locked anchors.

## 7. Evolutionary Changelog
*   **2026-06-22:** Initial Document Creation.
