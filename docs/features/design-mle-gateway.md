# Design Doc: Mission-Locked Execution (MLE) Gateway
**Status:** Draft
**Created:** 2026-06-07

## 1. Context and Scope
As agent swarms grow in complexity and autonomy, "Intent Ghosting" has emerged as a critical vulnerability. This occurs when a subagent diverges from its parent's verified intent, often by mimicking the parent's context or persona (Semantic Shadowing), and executes unauthorized tool calls. Current transport-layer bindings and binary handoffs are insufficient because they don't cryptographically "lock" the execution to a specific mission fragment.

The Mission-Locked Execution (MLE) Gateway is needed to ensure that every tool call, sub-delegation, and state mutation is anchored to a hardware-attested, immutable mission-root fragment at the point of issuance.

## 2. Goals & Non-Goals
* **Goals:**
    * Enforce cryptographic locking of all tool calls and delegations to mission-root intents.
    * Provide a "Zero-Trust Intent Chain" that persists throughout the agent lifecycle.
    * Neutralize "Intent Ghosting" and mimicry-based bypasses.
    * Integrate with hardware-attested (TPM) identity fragments for non-repudiable mission sovereignty.
* **Non-Goals:**
    * Replacing the Semantic Integrity Bridge (ISD). MLE provides the cryptographic lock, while ISD provides the semantic validation.
    * Managing the internal reasoning state of individual subagents.
    * Enforcing framework-specific syntax.

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Ensure that a "Database Specialist" subagent can only execute queries that are explicitly locked to the "Quarterly Audit" mission root.
* **The Happy Path (Tasks):**
    1. Parent agent issues a task delegation with a cryptographically signed "Mission Lock" token.
    2. Subagent attempts to call a `db_query` tool.
    3. MLE Gateway intercepts the tool call.
    4. MLE Gateway verifies that the `db_query` call is cryptographically locked to the active "Quarterly Audit" mission fragment.
    5. The gateway validates the hardware-attested lineage of the request.
    6. If the lock is valid and aligned with the mission root, the tool call is permitted.
    7. If a subagent attempts to "ghost" a different intent (e.g., `db_delete`), the MLE Gateway detects the missing/invalid lock and blocks the execution.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Agent Tool Call + Mission Lock] --> B[MLE Gateway Interceptor]
        B --> C[Cryptographic Lock Validator]
        C --> D[Mission-Root Lineage Checker]
        D --> E{Locked & Attested?}
        E -- Yes --> F[Tool Execution Permitted]
        E -- No --> G[Execution Blocked & Sovereignty Alert]
        H[Hardware-Attested Mission Root] --> D
        I[Intent-Fragment Registry] --> C
    ```
* **APIs / Interfaces:**
    * `mle.LockTask(missionToken, taskFragment) -> LockToken`: Issues a cryptographic lock for a specific task fragment.
    * `mle.ValidateExecution(lockToken, toolCall) -> Result`: Validates that a tool call is locked to the authorized mission.
* **Data Storage/State:**
    * **Mission Lock Registry:** High-speed, in-memory store for active mission locks, anchored to the hardware-attested session.
    * **Sovereignty Logs:** Immutable logs of all mission-lock validation events.

## 5. Alternatives Considered
* **Session-Bound Tokens:** Rejected because they only verify the *session*, not the specific *mission fragment*. Session tokens can be reused for unauthorized intents (Ghosting).
* **Manual HITL for Every Call:** Rejected because it doesn't scale for autonomous machine-speed swarms.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** MLE must be implemented at the gateway level and utilize hardware-bound keys (TPM/SEP) to prevent lock forging.
* **Observability:** Integrated with the "Mission-Root Lineage Tracer" in the UI for real-time visualization of intent-locked chains.

## 7. Evolutionary Changelog
* **2026-06-07:** Initial Document Creation.

### Update: 2026-06-08 - HAMM Integration & Manifest-Locked discovery
**Context:** Today's market sync revealed the emergence of "Hardware-Attested Mission Manifests" (HAMM) in Gemini CLI v0.38.0-alpha, requiring pre-declared tool intents.
**Architecture Adjustment:**
* Upgrading Section 4 to support HAMM-compliant lookups.
* Introducing a "Pre-Execution Manifest Validator" that cross-references tool calls against a TPM-signed manifest.
**Security Impact:** Prevents "Discovery-Phase Shadowing" by mandating that every possible capability be declared before sub-mission instantiation.
