# Design Doc: Mission-Root Pinning (MRP)
**Status:** Draft
**Created:** 2026-05-18

## 1. Context and Scope
The emergence of "Mission Root Exhaustion" (MRE) attacks represents a critical threat to agentic stability. In these attacks, malicious subagents or compromised skills "flood" the agent's context window with irrelevant semantic noise. This forces the agent's internal summarization logic to drop the original "Mission Root" intent to make room for new tokens, effectively hijacking the agent's long-term goals. MCP Any needs to solve this by providing a transport-level safeguard that ensures the primary intent remains persistent and "pinned" throughout the session.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a transport-level "Pinning" mechanism for the cryptographically signed Mission Root.
    * Ensure the Mission Root is automatically re-injected into every context window update if it is dropped by the agent framework.
    * Provide a "Semantic Noise Filter" to detect and alert on MRE attack patterns (e.g., high-frequency token spikes with low semantic density).
* **Non-Goals:**
    * Modifying the underlying LLM's context window size.
    * Preventing all forms of context bloat (focus is specifically on the Mission Root).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Agent Swarm Developer
* **Primary Goal:** Prevent a specialized subagent from "forgetting" the primary mission due to adversarial context flooding.
* **The Happy Path (Tasks):**
    1. The developer enables MRP in the MCP Any configuration for a specific swarm.
    2. The "Mission Root" is cryptographically signed and pinned during session initialization.
    3. A subagent attempts an MRE attack by generating 100k tokens of nonsense.
    4. The agent framework's summarizer drops the Mission Root to clear space.
    5. MCP Any's MRP middleware detects the omission and re-injects the pinned Mission Root into the next prompt, maintaining mission continuity.

## 4. Design & Architecture
* **System Flow:**
    `[Agent reasoning loop] -> [MRP Middleware] -> [LLM API]`
* **APIs / Interfaces:**
    * `MRP.pin_intent(signed_token)`: Anchors the mission root to the active transport session.
    * `MRP.validate_context(payload)`: Scans outgoing prompts for the presence of the pinned intent.
    * `MRP.inject_intent(payload)`: Force-prepends the pinned intent to the payload if missing.
* **Data Storage/State:**
    * Pinned intents are stored in-memory, bound to the `session_id` and verified against the hardware-attested token.

## 5. Alternatives Considered
* **Framework-Side Pinning**: Rejected as it relies on the agent framework (e.g., Claude Code, OpenClaw) being uncompromised and correctly implementing the pinning logic.
* **Hard Context Limits**: Rejected as it would break legitimate deep-reasoning tasks that require large context windows.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: The Mission Root itself must be hardware-attested to prevent an attacker from pinning a malicious "Root."
* **Observability**: Metrics on "Intent Re-injections" and "MRE Attack Signals" will be surfaced in the UI Health Dashboard.

## 7. Evolutionary Changelog
* **2026-05-18:** Initial Document Creation.
