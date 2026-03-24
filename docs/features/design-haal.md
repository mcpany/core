# Design Doc: Hardware-Attested Attention Locking (HAAL)
**Status:** Draft
**Created:** 2026-06-11

## 1. Context and Scope
As agent swarms scale, "Reasoning Entropy Exhaustion" (REE) has emerged as a significant threat. In REE attacks, subagents flood the parent agent's context window with high-entropy, semantically valid but mission-irrelevant reasoning traces. This effectively "blinds" the parent's attention mechanism, leading to the eviction of critical mission-root anchors and causing cognitive stall.

Hardware-Attested Attention Locking (HAAL) provides a cryptographic defense against REE. It allows the mission root to utilize hardware-bound headers to "lock" specific mission-critical fragments at the LLM attention layer, ensuring they cannot be evicted by subagent-injected noise.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a middleware that interfaces with LLM attention-locking APIs (e.g., Gemini's HAAL headers).
    * Provide a mechanism for the mission root to cryptographically sign and "lock" key intent fragments.
    * Neutralize "Reasoning Entropy Exhaustion" (REE) attacks in high-density swarms.
    * Integration with TPM/Secure Enclave for signing attention-lock requests.
* **Non-Goals:**
    * Manually managing LLM attention heads (handled by the model provider).
    * Providing general-purpose context window management (handled by the ContextEngine).
    * Blocking all high-entropy subagent noise (handled by L7SIH).

## 3. Critical User Journey (CUJ)
* **User Persona:** Mission-Root Orchestrator (e.g., Lead Claude Code Agent)
* **Primary Goal:** Maintain cognitive focus on the primary mission objectives despite aggressive noise injection from subagents.
* **The Happy Path (Tasks):**
    1. Orchestrator defines "Mission-Root Anchors" (critical intent fragments).
    2. Orchestrator requests HAAL locking for these anchors via the Universal Agent Bus.
    3. HAAL Middleware validates the request using the Orchestrator's hardware-attested identity.
    4. HAAL Middleware attaches TPM-signed attention-locking headers to all outbound LLM requests.
    5. The LLM provider recognizes the headers and prevents the locked fragments from being evicted during the reasoning cycle.
    6. Subagents attempt to flood the context window with REE noise, but the mission-root anchors remain pinned and accessible to the attention mechanism.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Orchestrator Fragment] --> B[HAAL Middleware]
        C[Hardware Identity Token] --> B
        B --> D[TPM/Enclave Signer]
        D --> E[Signed HAAL Header]
        E --> F[LLM Provider API]
        G[Subagent REE Noise] --> F
        F --> H[Attention Governance Layer]
        H -- Pin Fragments --> I[Reasoning Engine]
    ```
* **APIs / Interfaces:**
    * `haal.LockFragments(fragments []Fragment, identity HardwareIdentity) -> []Header`: Generates signed HAAL headers.
    * `haal.ValidateHeader(header Header) -> bool`: Verifies the integrity of an attention-lock header.
* **Data Storage/State:**
    * **Locked Fragment Registry:** A session-bound cache of fragments that have been successfully hardware-locked.

## 5. Alternatives Considered
* **Context-Window Pinning (CWP):** Rejected as CWP is a software-level hint that can be overridden by extreme entropy flooding. HAAL provides hardware-bound guarantees at the attention layer.
* **Dynamic Budgeting (ARE):** Complements HAAL but does not prevent attention-layer eviction once tokens are already consumed.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Attention-locking requests must be hardware-attested to prevent subagents from "locking" their own malicious fragments and evicting the mission root.
* **Observability:** Integrated with the "Context Attention Monitor" for real-time visualization of pinned vs. flooded fragments.

## 7. Evolutionary Changelog
* **2026-06-11:** Initial Document Creation. Evolving from Context-Window Pinning (2026-06-09) to support the hardware-attested HAAL standard.

### Update: 2026-06-19 - Hardware-Attested Context Protection
**Context:** Recent findings on deceptive context injection (Tracebit) confirm that natural language context is a primary attack vector for agent hijacking.
**Architecture Adjustment:**
* Extending HAAL to support **Context-File Integrity Attestation (CFIA)**.
* Mandating hardware-bound attention locking for any fragment ingested from a project-local configuration or context file.
**Security Impact:** Ensures that only user-verified, hardware-attested context can influence high-priority attention heads, neutralizing deceptive natural language injections.
