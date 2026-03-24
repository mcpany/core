# Design Doc: Policy-Bound Reasoning (PBR) Adapter
**Status:** Draft
**Created:** 2026-05-20

## 1. Context and Scope
As AI agents move toward deeper autonomy, "Tool Gating" (blocking a specific call) is becoming insufficient. Malicious instructions can manipulate an agent's reasoning loop to "hallucinate" authorization or bypass safety checks before a tool is ever called. The industry is moving toward "Policy-Bound Reasoning" (PBR), where models consult an immutable security policy during the reasoning process. MCP Any needs to act as the authoritative host for these "Policy Anchors," ensuring that cognitive governance is maintained even as tasks are delegated across heterogeneous agent swarms.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide a standardized interface for hosting and distributing "Policy Anchors" to connected agents.
    * Ensure that security policies are enforced at the pre-reasoning layer, neutralizing "Cognitive Bypasses."
    * Support cross-framework PBR enforcement (e.g., ensuring an OpenClaw subagent respects a Claude-originated policy).
* **Non-Goals:**
    * Replacing the internal reasoning engine of the model.
    * Real-time modification of model weights.

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise Security Architect
* **Primary Goal:** Prevent a coding agent from even "thinking" about exfiltrating secrets, regardless of whether it has access to a `curl` tool.
* **The Happy Path (Tasks):**
    1. The architect defines a "Cognitive Policy Anchor" in MCP Any (e.g., "Never reason about external network connections for this mission").
    2. When the primary agent boots, it ingests the hardware-attested PBR Anchor from MCP Any.
    3. During its chain-of-thought, the agent encounters a task that might require network access.
    4. The PBR mechanism evaluates the reasoning path against the Anchor *before* generation.
    5. The model prunes the unauthorized reasoning path, remaining within the defined safety boundary.

## 4. Design & Architecture
* **System Flow:**
    `[Policy Store] -> [PBR Adapter] -> [Agent Prompt/System Context Injection]`
* **APIs / Interfaces:**
    * `PBR.get_policy_anchor(mission_root_id)`: Returns the cryptographically signed policy for a specific mission.
    * `PBR.validate_reasoning_step(step_content, anchor_id)`: (Optional) Server-side validation for frameworks that support external reasoning checks.
* **Data Storage/State:**
    * Policy Anchors are stored as immutable, content-addressed fragments (CAC) in the hardware-attested registry.

## 5. Alternatives Considered
* **Reactive Tool Blocking**: Rejected as it only catches the *action*, not the *intent* or *hallucination* that led to it.
* **System Prompt Hardening**: Insufficient, as system prompts can be "diluted" or bypassed in deep context windows.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: PBR is the ultimate "Pre-Thought" gate, reducing the attack surface to the integrity of the Policy Anchor itself.
* **Observability**: "Policy Pruning" events and "Cognitive Violations" are logged and surfaced in the Reasoning Alignment Visualizer.

## 7. Evolutionary Changelog
* **2026-05-20:** Initial Document Creation.
* **2026-05-21:** Added **Temporal Reasoning Attestation (TRA)** requirement to the PBR Adapter to neutralize "Reasoning Timing Attacks." Updated Section 4 to include hardware-attested monotonic timestamps for all policy anchor handoffs.
