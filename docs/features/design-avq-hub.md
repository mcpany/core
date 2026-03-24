# Design Doc: Autonomous Verification Quorum (AVQ) Hub
**Status:** Draft
**Created:** 2026-06-02

## 1. Context and Scope
The **Autonomous Verification Quorum (AVQ) Hub** is a core security component designed to solve the "Mean Time to Consensus" (MTTC) bottleneck in high-density agent swarms. In autonomous workflows, high-stakes tasks (e.g., recursive filesystem deletions, production API key rotation) currently require expensive and slow human-in-the-loop (HITL) approvals or suffer from uncoordinated, hallucinatory tool calls. The AVQ Hub facilitates a **hardware-attested, multi-agent consensus** on high-risk actions, allowing a swarm to authorize its own actions within cryptographically signed mission boundaries.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Reduce the Mean Time to Consensus (MTTC) in 10+ agent swarms to under 500ms.
    *   Enforce mandatory TPM-bound signatures for every agent participating in a quorum.
    *   Provide a standardized "Consensus Token" for tool middleware to verify authorization.
    *   Support "Dynamic Quorum Scaling" based on the real-time risk score of the tool call.
*   **Non-Goals:**
    *   Directly executing the tool calls (this is handled by the Tool Execution Engine).
    *   Replacing human oversight for "Mission-Root" changes (these still require global HITL).
    *   Solving for non-A2A compliant agents (AVQ requires hardware-attested NHI).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Heterogeneous Swarm Orchestrator (e.g., Local LLM Swarm Orchestrator)
*   **Primary Goal:** Authorize a high-risk filesystem commit across 3 parallel agents without manual user approval.
*   **The Happy Path (Tasks):**
    1.  Agent A proposes a high-risk `fs_commit` to the AVQ Hub.
    2.  AVQ Hub identifies the "Mission-Root" and calculates the required quorum (3 agents).
    3.  AVQ Hub broadcasts a "Verification Challenge" to specialized Auditor Agents (B and C).
    4.  Auditor Agents B and C perform independent semantic validation of the reasoning trace and sign the challenge using their hardware-bound (TPM) session keys.
    5.  AVQ Hub aggregates the signatures and issues a **Hardware-Attested Quorum (HAQ) Token**.
    6.  The Tool Execution Engine receives the HAQ Token, verifies the mission-lineage, and executes the commit.

## 4. Design & Architecture
*   **System Flow:**
    ```mermaid
    graph TD
        A[Proposing Agent] -->|Task Proposal + TPM Signature| B[AVQ Hub]
        B -->|Quorum Challenge| C[Auditor Agent 1]
        B -->|Quorum Challenge| D[Auditor Agent 2]
        C -->|TPM-Signed Attestation| B
        D -->|TPM-Signed Attestation| B
        B -->|HAQ Token| E[Tool Execution Engine]
    ```
*   **APIs / Interfaces:**
    *   `POST /avq/v1/propose`: Propose a task for quorum validation.
    *   `POST /avq/v1/attest`: Submit a hardware-attested signature for a pending challenge.
    *   `GET /avq/v1/status/{challenge_id}`: Poll for the status of a quorum.
*   **Data Storage/State:**
    *   Quorum state is managed in a specialized, intent-sealed shard within the Shared KV Store (Blackboard), ensuring non-repudiation and persistence.

## 5. Alternatives Considered
*   **Centralized Orchestrator Approvals:** Rejected due to the "Single Point of Failure" and the latency bottleneck in massive swarms.
*   **Soft Consensus (No TPM):** Rejected due to the high risk of "Reasoning Mirroring" and identity-spoofing in heterogeneous environments.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** The AVQ Hub utilizes hardware-bound (TPM/SEP) keys to ensure that every participant in the quorum is an authorized, resident agent of the local machine. This prevents remote "Sybil" attacks where a rogue cloud agent attempts to spoof a local swarm quorum.
*   **Observability:** All quorum events, signatures, and time-to-consensus (TTC) metrics are logged to the `mcpany` audit log with cryptographic lineage.

## 7. Evolutionary Changelog
*   **2026-06-02:** Initial Document Creation.
