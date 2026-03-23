# Design Doc: Mission Anchor Kernel (MAK)
**Status:** Draft
**Created:** 2026-05-10

## 1. Context and Scope
As AI agent swarms grow in depth and complexity, they increasingly suffer from "Reasoning Drift"—a state where subagents lose track of the primary mission boundaries and begin performing out-of-scope actions. Current security models rely on the agent's internal reasoning to respect boundaries, which is fundamentally fragile. The Mission Anchor Kernel (MAK) moves intent enforcement from the cognitive layer to the transport layer.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Provide a non-bypassable, transport-layer enforcement engine for mission goals.
    *   Bind every tool call to a cryptographically signed "Mission Manifest."
    *   Detect and halt subagent actions that semantically diverge from the root mission.
*   **Non-Goals:**
    *   Modifying the internal weights or reasoning logic of the LLM.
    *   Implementing a full task-specific linter (MAK handles high-level intent).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Enterprise Security Architect
*   **Primary Goal:** Ensure that a 10-agent recursive research swarm never makes unauthorized network calls, even if individual agents hallucinate a need for them.
*   **The Happy Path (Tasks):**
    1.  Architect defines a "Mission Manifest" containing the root goal and permitted resource scopes.
    2.  MCP Any signs the manifest and attaches it to the primary agent's session.
    3.  A level-4 subagent attempts to call the `aws:execute_shell` tool for a non-research task.
    4.  The MAK intercepts the call, compares the request against the signed manifest using a fast semantic validator, and rejects the call.
    5.  The architect receives a "Mission Violation" alert in the dashboard.

## 4. Design & Architecture
*   **System Flow:**
    ```mermaid
    graph TD
        Agent[Subagent] -->|Tool Request| MAK[Mission Anchor Kernel]
        MAK -->|Verify Signature| Auth[Manifest Authenticator]
        MAK -->|Semantic Check| Policy[Intent Policy Engine]
        Policy -->|Approved| Tool[MCP Tool/Server]
        Policy -->|Violation| Dashboard[Security Dashboard]
    ```
*   **APIs / Interfaces:**
    *   `NewMission(Manifest) -> MissionID`: Initializes a new anchored mission.
    *   `MAK.Intercept(Request, MissionID) -> Result`: Internal middleware interface for tool validation.
*   **Data Storage/State:**
    *   Manifests are stored as append-only logs in the **Shared KV Store** with hardware-bound encryption keys.

## 5. Alternatives Considered
*   **Prompt-Based Guardrails**: Rejected because they are susceptible to "jailbreaking" and reasoning hallucinations.
*   **Static RBAC**: Rejected because agent missions are dynamic and require context-aware enforcement that static tokens cannot provide.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** MAK uses "Positive Attestation"—no action is allowed unless it is explicitly derived from the signed mission root.
*   **Observability:** All MAK interceptions are logged with "Semantic Drift Scores" to help architects tune their manifests.

## 7. Evolutionary Changelog
*   **2026-05-10:** Initial Document Creation.
