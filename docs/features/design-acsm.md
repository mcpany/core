# Design Doc: Action-Chain Sovereignty Monitor (ACSM)
**Status:** Draft
**Created:** 2026-07-08

## 1. Context and Scope
With the official experimental rollout of Claude Code "Agent Teams" and the rise of coordinated "Hivenet" swarm attacks, the security frontier for AI agents has moved from protecting individual tool calls to governing the **entire sequence of actions** across a mesh of teammates. A compromised agent may not trigger any single-point alarms but can execute a "low-and-slow" multi-point breach that is only detectable through cross-agent action-chain analysis.

ACSM is the evolution of MCP Any's Collective Swarm Anomaly Detection (CSAD). It moves from a reactive, per-call validation model to an active, sequence-aware monitor that ensures every action-chain remains semantically bound to the mission-root intent and doesn't exhibit the high-entropy behavior of a coordinated attack.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Maintain a real-time, hardware-attested graph of action-chains across all connected agents.
    *   Detect and interdict action sequences that diverge from the verified mission-root manifest.
    *   Provide sub-millisecond anomaly detection for coordinated "Hivenet" probes.
    *   Support hardware-bound "Sovereignty Proofs" for entire multi-agent workflows.
*   **Non-Goals:**
    *   Replacing individual tool call validation (ACSM works *alongside* per-call security).
    *   Governing human users (ACSM is focused exclusively on autonomous agent action-chains).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Enterprise Security Architect
*   **Primary Goal:** Detect and block a "Hivenet" attack where 3 different Claude teammates each perform "benign" discovery actions that, when combined, map the internal network for exfiltration.
*   **The Happy Path (Tasks):**
    1.  The Team Lead initiates a mission with a hardware-attested manifest.
    2.  ACSM begins tracking the "Action-Chain Root" for this mission.
    3.  Teammate A calls `ls /tmp`. ACSM records the sequence.
    4.  Teammate B calls `netstat -an`. ACSM detects a "Discovery Sequence" forming.
    5.  Teammate C calls `curl internal-registry.local`. ACSM identifies the aggregate entropy exceeds the "Mission Discovery Budget".
    6.  ACSM triggers an immediate interdiction, revoking all teammate capabilities for that mission root.

## 4. Design & Architecture
*   **System Flow:**
    ```mermaid
    graph TD
        A[Agent Tool Call] --> B[ELIG Middleware]
        B --> C{ACSM Sequence Engine}
        C --> D[Mission Graph Store]
        C --> E[Anomaly Scoring Model]
        E -->|High Entropy| F[Interdiction Hub]
        E -->|Safe| G[Upstream MCP Server]
        D -->|Sequence Match| E
    ```
*   **APIs / Interfaces:**
    *   `POST /v1/acsm/verify-sequence`: Internal endpoint for ELIG to check if the next step in the chain is authorized.
    *   `GET /v1/acsm/chain-topology`: UI endpoint for the Action-Chain Sovereignty Monitor visualization.
*   **Data Storage/State:**
    *   Utilizes the Shared KV Store (Blackboard) with a specialized `action_chains` namespace.
    *   Action-chains are stored as cryptographically hash-chained sequences, pinned to the mission-root.

## 5. Alternatives Considered
*   **Stateless Rate Limiting:** Rejected because coordinated "Hivenets" operate specifically to stay under per-agent rate limits.
*   **Centralized Model-Based Review:** Rejected due to the prohibitive latency tax for real-time inter-agent coordination.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** ACSM itself must be hardware-attested. Any tampering with the Mission Graph Store triggers a mesh-wide lockdown.
*   **Observability:** Integrated with the Action-Chain Sovereignty Monitor UI for real-time "Entropy Heatmaps."

## 7. Evolutionary Changelog
*   **2026-07-08:** Initial Document Creation (Evolution of CSAD).
