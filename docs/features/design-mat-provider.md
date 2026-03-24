# Design Doc: Mission Alignment Token (MAT) Provider
**Status:** Draft
**Created:** 2026-05-16

## 1. Context and Scope
As agent swarms increase in recursive depth, "Contextual Decay" becomes a significant failure mode. Research indicates that every layer of delegation introduces a roughly 12% drift from the original "Mission Root" intent. This lead to specialized subagents pursuing sub-goals that may contradict the user's primary objective.

The MAT Provider is a middleware designed to counteract this decay. It injects hardware-attested "Alignment Tokens" into every context fragment handed between agents. These tokens act as a "North Star," providing a high-weight semantic signal that the reasoning engine must prioritize. By binding these tokens to the hardware-protected Mission Root, we ensure that the primary objective remains the dominant feature of the reasoning trace, regardless of swarm depth.

## 2. Goals & Non-Goals
* **Goals:**
    * Automatically inject hardware-attested alignment tokens into inter-agent context handoffs.
    * Use IBHI (Intent-Bound Hardware Isolation) to protect the root intent used for token generation.
    * Provide a "Mission Persistence Score" to users to visualize alignment across the swarm.
    * Mandate that agents attest to the MAT before executing high-privilege tools.
* **Non-Goals:**
    * Rewriting the agent's internal monologue (MAT is an external context feature).
    * Preventing all forms of hallucination (MAT specifically targets goal drift).

## 3. Critical User Journey (CUJ)
* **User Persona:** Swarm Orchestrator
* **Primary Goal:** Maintain strict adherence to a "Primary Directive" (e.g., "Do not modify production databases") through 10+ layers of subagent delegation.
* **The Happy Path (Tasks):**
    1. The User defines a Mission Root and locks it into MCP Any's hardware vault.
    2. A Parent Agent delegates a task to a Sub-Specialist.
    3. The MAT Provider intercepts the handoff and generates an Alignment Token based on the hardware-locked Root.
    4. The Sub-Specialist receives the context with the MAT at the top of its prompt.
    5. When the Sub-Specialist attempts a tool call, MCP Any verifies that the MAT is present and un-modified in the agent's recent context.
    6. If the agent has drifted (e.g., removed the MAT from context), the tool call is blocked, and an "Alignment Error" is raised.

## 4. Design & Architecture
* **System Flow:**
    [Hardware Root] -> [MAT Generator] -> [Context Fragment] -> [Agent Reasoning] -> [Tool Call Validator]
    1. MAT Generator periodically pulls the Root Intent from the TPM.
    2. It signs the intent with a session-nonce to create the MAT.
    3. Context Bridge appends the MAT to all outbound context handoffs.
* **APIs / Interfaces:**
    * `GenerateMAT(intent_id) -> alignment_token`
    * `VerifyAlignment(context_blob) -> bool`
* **Data Storage/State:**
    * The Root Intent is stored in a TPM-sealed segment.
    * MATs are ephemeral and bound to specific agent sessions.

## 5. Alternatives Considered
* **Static Context Pinning:** Rejected because static pins are easily ignored by LLMs during long reasoning chains. MAT requires active attestation at the tool-call boundary.
* **Prompt Injection Defense (PID):** MAT is complementary to PID, but focuses on maintaining parent intent rather than blocking external malicious input.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The MAT is cryptographically bound to the hardware root. Any tampering with the token invalidates the agent's ability to call tools.
* **Observability:** "Alignment Decay" is tracked as a telemetry metric, allowing users to identify which sub-specialists are most prone to mission drift.

## 7. Evolutionary Changelog
* **2026-05-16:** Initial Document Creation.
