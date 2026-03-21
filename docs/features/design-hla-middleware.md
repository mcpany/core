# Design Doc: Hardware-Locked Attention (HLA) Middleware
**Status:** Draft
**Created:** 2026-06-25

## 1. Context and Scope
As agents operate in high-density swarms, they are increasingly susceptible to "Chain-of-Thought Poisoning." Malicious subagents can inject mission-divergent reasoning fragments into the shared context, attempting to evict core instructions from the LLM's attention window. Transport-layer security is no longer enough; we must secure the **cognitive path** by ensuring that critical mission anchors remain "pinned" in the model's attention.

## 2. Goals & Non-Goals
* **Goals:**
    * Utilize hardware-bound attention-locking headers to cryptographically "pin" mission-critical intent fragments.
    * Prevent eviction of core instructions by high-entropy noise injected by subagents.
    * Provide a verifiable audit trail of attention-layer prioritization.
* **Non-Goals:**
    * Modifying the internal weights or attention mechanisms of the model (HLA acts on the prompt construction and provider-specific attention hints).
    * Providing a general-purpose prompt optimizer.

## 3. Critical User Journey (CUJ)
* **User Persona:** Mission Root Administrator
* **Primary Goal:** Ensure that a specialized subagent cannot ignore its core constraints, even if it encounters conflicting data.
* **The Happy Path (Tasks):**
    1. The administrator defines "Attention Anchors" for the mission (e.g., "Do not exfiltrate PII").
    2. HLA Middleware signs these anchors with the hardware-attested mission key.
    3. During the tool-call loop, HLA injects these signed anchors using provider-specific headers (e.g., `x-attention-lock`).
    4. The LLM prioritizes these locked fragments during reasoning.
    5. HLA monitors the output to ensure the reasoning remains aligned with the locked anchors.

## 4. Design & Architecture
* **System Flow:**
    `Agent -> HLA Middleware (Injects Attention Locks) -> LLM Provider -> Reasoning Output`
* **APIs / Interfaces:**
    * `HLA.v1.LockFragment(fragment_id, priority_level) -> signed_header`
    * `HLA.v1.VerifyAlignment(reasoning_trace) -> alignment_score`
* **Data Storage/State:**
    * Attention policies are stored in the **Mesh-Resident Policy Synthesizer**.
    * Signature keys are bound to the host TPM.

## 5. Alternatives Considered
* **Software-based Prompt Repetition:** Rejected because it consumes excessive tokens and can be bypassed by long-context poisoning.
* **Model-level Fine-tuning:** Rejected as it lacks the dynamic flexibility required for diverse agent swarms.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Only the mission root can issue attention-lock signatures.
* **Observability:** Visualized via the **Visual Attention Dashboard** (P1 feature).

## 7. Evolutionary Changelog
* **2026-06-25:** Initial Document Creation.
