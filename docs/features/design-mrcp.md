# Design Doc: Mission-Root Continuity Provider (MRCP)
**Status:** Draft
**Created:** 2026-06-21

## 1. Context and Scope
With the release of OpenClaw v3.1.2 and the shift toward "Mission-Locked Execution" (MLE), AI agent sessions are becoming longer-running and more decentralized. Long-running missions are susceptible to "Cognitive Stall" due to system restarts, teammate rotations, or framework handoffs, which often lead to context loss and mission failure.

The **Mission-Root Continuity Provider (MRCP)** is an infrastructure layer for MCP Any that facilitates hardware-locked mission resumption. It ensures that the cryptographically signed reasoning-path and associated state fragments persist across the entire mission lifecycle, regardless of the execution environment or active agent framework.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Facilitate sub-100ms mission resumption across system restarts.
    *   Maintain hardware-attested continuity of the reasoning-path.
    *   Provide a secure "Resumption Token" bound to the hardware-attested session.
    *   Synchronize continuity signals between disparate frameworks (OpenClaw, AutoGen, Claude Code).
*   **Non-Goals:**
    *   Replacing local agent memory management.
    *   Providing global context storage for all agents (handled by the Blackboard).
    *   Auto-restarting failed agent processes (handled by the orchestrator).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Enterprise Swarm Operator
*   **Primary Goal:** Resume a multi-day data analysis mission after a scheduled server maintenance restart without losing reasoning progress.
*   **The Happy Path (Tasks):**
    1.  The primary mission agent performs a checkpoint via MRCP, receiving a hardware-attested Resumption Token.
    2.  The server restarts.
    3.  Upon reboot, the orchestrator invokes MCP Any with the Resumption Token.
    4.  MRCP validates the token against the TPM and restores the signed mission-root intent and last valid reasoning fragment.
    5.  The agent framework ingests the restored continuity state and resumes reasoning from the last checkpoint.

## 4. Design & Architecture
*   **System Flow:**
    ```mermaid
    graph TD
        Agent[Agent Framework] -->|Checkpoint| MRCP[MRCP Hub]
        MRCP -->|Sign| TPM[Hardware Enclave/TPM]
        MRCP -->|Store| DB[Continuity Store - SQLite]

        Restart((System Restart))

        Orchestrator -->|Resume + Token| MRCP
        MRCP -->|Verify| TPM
        MRCP -->|Retrieve| DB
        MRCP -->|Continuity State| Agent
    ```
*   **APIs / Interfaces:**
    *   `POST /mission/checkpoint`: Generates a hardware-attested resumption token.
    *   `POST /mission/resume`: Validates token and returns continuity state.
    *   `Header: x-mcp-continuity-token`: Used for all inter-teammate coordination.
*   **Data Storage/State:**
    *   Continuity state is stored in a dedicated SQLite sidecar, indexed by hardware-bound Mission IDs.
    *   Reasoning fragments are hash-chained to ensure integrity during resumption.

## 5. Alternatives Considered
*   **Stateless Resumption**: Rejected as it relies on agent frameworks to perfectly reconstruct context, leading to high hallucination risk and token waste.
*   **Blackboard-Only Persistence**: Rejected as it lacks the hardware-attestation requirements needed for Mission-Locked Execution (MLE) compliance.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** Resumption tokens are time-bound and cryptographically linked to the specific hardware Inode of the mission manifest.
*   **Observability:** MRCP logs all resumption events with monotonic counters to detect "Replay-as-Resumption" attacks.

## 7. Evolutionary Changelog
*   **2026-06-21:** Initial Document Creation.

### Update: 2026-06-25 - Monotonic Lineage Attestation
**Context:** Today's market sync revealed "Snapshot Corruption" risks in AMR and "Identity Leakage via Process Environment."
**Architecture Adjustment:**
*   Integrating **Monotonic Mission Lineage (MML)** into the resumption flow.
*   Resumption tokens now require a TPM-signed monotonic counter to prevent replay and leakage.
*   Deprecating plain environment variables for identity transport in favor of kernel-bound HLES buffers.
**Security Impact:** Prevents subagents from "squatting" on stale resumption tokens and ensures environmental sovereignty for headless handoffs.

### Update: 2026-06-23 - Recursive Mission-Root Attestation & AIS Integration
**Context**: Today's market sync revealed "Governance Gaps" in headless handoffs and "Teammate Mailbox Splicing" (CVE-2026-81042).
**Architecture Adjustment**:
*   Mandating **Recursive Mission-Root Attestation (RMRA)** for all process-based subagent handoffs in Section 4.
*   Integrating the **Active Intent Sanitizer (AIS)** into the coordination bus to block cross-channel side-channel exfiltration.
**Security Impact**: Neutralizes intent drift during deep sub-process chains and prevents mailbox splicing via higher-dimensional behavioral anchoring.
