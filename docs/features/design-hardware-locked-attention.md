# Design Doc: Hardware-Locked Attention (HLA) Provider
**Status:** Draft
**Created:** 2026-06-25

## 1. Context and Scope
As AI models move toward 1M+ token context windows (Gemini, Claude), a new attack vector has emerged: **Context-Window Flooding (CWF)**. Malicious tool outputs or "deceptive context" (e.g., a weaponized `GEMINI.md` file) can provide high-entropy noise that "evicts" the system instructions or mission-root intents from the model's active attention.

MCP Any needs to ensure that core mission constraints remain "pinned" at the attention-head level, regardless of the volume of downstream tool output. The HLA Provider acts as a cryptographically bound attention-governance layer.

## 2. Goals & Non-Goals
* **Goals:**
    * Cryptographically "pin" mission-root fragments in the model's attention window.
    * Automatically summarize or truncate tool outputs that exceed "Attention-Density" safety thresholds.
    * Provide hardware-attested proof that the model is still "attending" to the user's primary intent.
* **Non-Goals:**
    * Replacing the model's internal attention mechanism.
    * Modifying the model weights or architecture.
    * Managing long-term memory (this is handled by the ContextEngine).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious DevSecOps Engineer
* **Primary Goal:** Prevent a subagent from being "social-engineered" into ignoring its security constraints via a 500k-token "noise" injection.
* **The Happy Path (Tasks):**
    1. The user defines a "Mission Root" with strict security anchors.
    2. The HLA Provider generates a hardware-attested "Attention Mask" for these anchors.
    3. A subagent encounters a malicious file containing 800k tokens of "Ignore previous instructions" noise.
    4. The HLA Provider detects the "Attention-Density" spike and prioritizes the "Pinned" anchors in the prompt construction.
    5. The model retains the mission-critical constraints and ignores the malicious noise.

## 4. Design & Architecture
* **System Flow:**
    `Agent -> [HLA Middleware] -> [Attention-Density Firewall] -> Model`
    The HLA Middleware injects specialized headers (compatible with Gemini/Claude ARE extensions) that instruct the provider's infrastructure to prioritize specific tokens.
* **APIs / Interfaces:**
    * `PinAttention(fragment_id, hardware_token)`: Locks a specific context fragment.
    * `GetAttentionHeatmap()`: Returns the real-time distribution of the model's focus.
* **Data Storage/State:**
    Attention masks are stored in the Shared KV Store (Blackboard) and bound to the hardware-attested session.

## 5. Alternatives Considered
* **Recursive Summarization:** Rejected because summarizing the entire window adds too much latency and can still lose nuances of the intent.
* **Hard Token Limits:** Rejected as it breaks valid use cases for large-context processing.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The HLA headers must be cryptographically signed. An unauthenticated subagent cannot "pin" its own instructions to hijack the model.
* **Observability:** Integrated with the **Visual Attention Dashboard** to provide real-time heatmaps to the user.

## 7. Evolutionary Changelog
* **2026-06-25:** Initial Document Creation.
