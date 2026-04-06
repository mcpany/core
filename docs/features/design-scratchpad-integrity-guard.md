# Design Doc: Project-Local Scratchpad Guard (PLSG)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
With the introduction of shared team workspaces (e.g., Claude Code's `.scratchpad` files), agent swarms now have a high-bandwidth channel for coordinating state. However, this channel lacks cognitive security. Specialist subagents can "stitch" together sensitive fragments of parent context and write them to the scratchpad, where they can be exfiltrated or misread by other teammates.

The Project-Local Scratchpad Guard (PLSG) is required to provide real-time semantic monitoring and redaction for all writes to shared team workspaces, ensuring mission-root privacy and preventing "Context-Stitching."

## 2. Goals & Non-Goals
* **Goals:**
    * Intercept and perform semantic analysis on all writes to project-local scratchpad files.
    * Implement **Reasoning-Aware Redaction (RAR)** to remove mission-root intents and PII from shared state.
    * Provide a kernel-level **Atomic Scratchpad Arbiter** to prevent coordination race conditions.
    * Maintain a cryptographically signed audit log of all workspace mutations.
* **Non-Goals:**
    * Replacing general-purpose file-watching services.
    * Managing scratchpad content outside the authorized agent team scope.
    * Providing long-term archival for scratchpad history.

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Prevent a "Researcher" subagent from leaking private database schema details (from the parent context) into the shared `.scratchpad`.
* **The Happy Path (Tasks):**
    1. Parent agent spawns a Researcher subagent to analyze a data fragment.
    2. Researcher subagent attempts to write a summary including the database schema to `.scratchpad`.
    3. PLSG intercepts the file-write operation.
    4. The RAR Engine identifies the schema details as "Sensitive Parent Context" and redacts them.
    5. The Atomic Arbiter ensures no other teammate is writing to the same block.
    6. The redacted summary is committed to the scratchpad.
    7. An "Intent-Redacted" signal is sent back to the parent agent's audit log.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Subagent] -->|Write Request| B[PLSG Middleware]
        B --> C{RAR Engine}
        C -->|Sensitive Content| D[Redaction Layer]
        C -->|Safe Content| E[Atomic Arbiter]
        D --> E
        E -->|Atomic Commit| F[.scratchpad file]
        B --> G[Sovereignty Audit Log]
    ```
* **APIs / Interfaces:**
    * `plsg.InterceptWrite(filePath, content, agentID) -> RedactedContent`: Main interception hook.
    * `plsg.AcquireLock(filePath, missionToken) -> LockID`: Atomic locking for coordination.
    * `plsg.ReleaseLock(lockID)`: Cleanup after write completion.
* **Data Storage/State:**
    * **Sensitivity Manifest:** A real-time index of "Mission-Root" fragments that must be redacted if they appear in tool outputs.
    * **Lock Registry:** In-memory store of active workspace locks.

## 5. Alternatives Considered
* **Manual Redaction by Agents:** Rejected because compromised or hallucinating agents cannot be trusted to redact their own output.
* **Encrypted Scratchpads:** Rejected because teammates need to read the scratchpad to coordinate; encryption only prevents host-level access, not inter-agent leakage.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The RAR Engine uses the same hardware-attested mission-root tokens to identify what fragments belong to the parent vs. the child.
* **Observability:** Integrated with the "Scratchpad Integrity Dashboard" for real-time visualization of redaction events.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
