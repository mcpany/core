# Design Doc: Agent-Facing Security Anchors (AFSA)
**Status:** Draft
**Created:** 2026-04-10

## 1. Context and Scope
As AI agents move toward full autonomy, the boundary between "User Intent" and "Agent Action" becomes increasingly blurred. Traditional security models focus on external enforcement (blocking a tool call after it's been made). However, the emergence of "Agent-Facing" defense (e.g., SlowMist's security guides for agents) suggests that agents should be active participants in their own security posture.

MCP Any needs to solve the "Semantic Alignment" gap, where an agent understands *what* it can do but doesn't have a structured, immutable way to reason about *why* certain actions are restricted. AFSA provides "Self-Hardening" manifests that are injected into the agent's reasoning loop, making Zero-Trust constraints a core reasoning primitive.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide structured, hardware-attested security manifests for agent ingestion.
    * Enable agents to self-police reasoning paths before tool execution.
    * Standardize the communication of security "intent" from the infrastructure to the agent.
    * Support "GC-Immune" pinning of these anchors in the model's attention window.
* **Non-Goals:**
    * Replacing the hardware-enforced Policy Firewall (AFSA is a first-line reasoning defense).
    * Modifying the internal weights or architecture of the LLM.

## 3. Critical User Journey (CUJ)
* **User Persona:** Autonomous Agent Swarm Orchestrator
* **Primary Goal:** Ensure a swarm of specialist agents remains bound by organizational safety policies without constant human supervision.
* **The Happy Path (Tasks):**
    1. The Orchestrator initiates a new mission.
    2. MCP Any generates a mission-specific AFSA manifest based on the user's security profile.
    3. The primary agent ingests the AFSA manifest as a "Reasoning Anchor."
    4. During task execution, a subagent proposes an action that would violate a policy (e.g., accessing a `.env` file).
    5. The primary agent, guided by the AFSA, recognizes the violation in its "Internal Monologue" and interdicts the subagent before the request even reaches the tool gateway.
    6. The event is logged as a "Self-Correction" event in the MCP Any dashboard.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Mission Root] --> B[AFSA Provider]
        B --> C[Hardware-Attested Manifest]
        C --> D[Agent Reasoning Loop]
        D --> E{Policy Check}
        E -->|Aligns| F[Tool Execution]
        E -->|Diverges| G[Self-Correction]
    ```
* **APIs / Interfaces:**
    * `GetSecurityAnchor(mission_id)`: Retrieves the cryptographically signed manifest.
    * `SubmitReasoningTrace(trace)`: (Optional) Allows the agent to submit its CoT for side-channel validation against the anchor.
* **Data Storage/State:** Anchors are stored in the "Sovereign Anchor Store," bound to the hardware-attested mission session.

## 5. Alternatives Considered
* **Static System Prompts:** Rejected because they are prone to "Instruction Eviction" in long-context models and can be easily bypassed by prompt injection. AFSA is treated as a high-priority "Semantic Fragment" that is hardware-pinned.
* **Purely External Enforcement:** Rejected because it doesn't solve "Hallucination-Driven Exploits" where the agent thinks an action is safe and continues to retry, wasting tokens and reasoning effort.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The manifest itself must be signed by the hardware root of trust (TPM) to prevent tampering.
* **Observability:** Self-correction events are prioritized in the security telemetry stream, providing high-signal data for RLHF/fine-tuning.

## 7. Evolutionary Changelog
* **2026-04-10:** Initial Document Creation.
