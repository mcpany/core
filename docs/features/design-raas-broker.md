# Design Doc: Reflection-as-a-Service (RaaS) Broker
**Status:** Draft
**Created:** 2026-07-09

## 1. Context and Scope
As AI agent swarms grow in depth and autonomy, the risk of "Internal Corruption"—where a specialist agent logic-loops or hallucinates and then commits corrupted state to the shared Blackboard—becomes a critical failure mode. OpenClaw v3.5.0 introduced Agent Reflection Quorums (ARQ) to mitigate this by requiring peer validation of reasoning traces. However, implementing ARQs within each agent framework leads to high MTTC (Mean Time to Coordinate) and fragmented security policies.

MCP Any will implement a centralized **Reflection-as-a-Service (RaaS) Broker**. This service acts as the authoritative arbitrator for state-commit reflections, providing a framework-neutral bus where reasoning traces from Claude, OpenClaw, or AutoGen agents can be cross-validated by independent peer "Auditors" before any global state change is finalized.

## 2. Goals & Non-Goals
* **Goals:**
    * Orchestrate multi-agent quorums (ARQs) for high-stakes state commits.
    * Provide a standardized format for reasoning-trace ingestion across different frameworks.
    * Dynamically select "Auditor" agents based on task-expertise and reputation.
    * Implement "Optimistic Reflection" to reduce the latency tax on state commits.
    * Enforce mission-root sovereignty by ensuring Auditors are independent of the specialist's immediate delegation branch.
* **Non-Goals:**
    * Performing the actual reasoning/LLM calls (delegated to mesh agents).
    * Providing long-term archival of reasoning traces (traces are ephemeral for the duration of the quorum).
    * Replacing the Blackboard (RaaS is a gatekeeper for it).

## 3. Critical User Journey (CUJ)
* **User Persona**: Enterprise Swarm Architect
* **Primary Goal**: Ensure that high-trust system modifications proposed by autonomous subagents are peer-reviewed for logic errors without manual human intervention.
* **The Happy Path (Tasks):**
    1. A "Specialist Agent" (e.g., a DevOps Specialist) completes a task and generates a proposed state change (e.g., a K8s manifest edit) along with its reasoning trace.
    2. The Specialist sends a `PROPOSE_COMMIT` request to the MCP Any Gateway.
    3. The **RaaS Broker** intercepts the request and identifies it as a "High-Impact" action requiring a quorum.
    4. The Broker selects two "Auditor Agents" (e.g., a Security Auditor and a SRE Auditor) from the active mesh.
    5. The Broker sends the reasoning trace and proposed state change to the Auditors in parallel.
    6. Auditors return cryptographically signed `APPROVE` tokens after successful reflection.
    7. The RaaS Broker aggregates the votes and triggers the final commit to the Blackboard.
    8. The Specialist receives a `COMMIT_SUCCESS` signal.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    sequenceDiagram
        participant S as Specialist Agent
        participant B as RaaS Broker
        participant A1 as Auditor Agent 1
        participant A2 as Auditor Agent 2
        participant BB as Blackboard (KV Store)

        S->>B: Propose Commit (State + Trace)
        B->>B: Identify Quorum Requirements
        par Broker to Auditors
            B->>A1: Request Reflection (Trace)
            B->>A2: Request Reflection (Trace)
        end
        A1-->>B: Vote: Approve (Signed)
        A2-->>B: Vote: Approve (Signed)
        B->>B: Validate Quorum (2/2)
        B->>BB: Execute Commit
        BB-->>B: Success
        B-->>S: Commit Confirmed
    ```
* **APIs / Interfaces:**
    * `POST /v1/reflection/propose`: Ingests a state change + reasoning trace. Returns a `ReflectionID`.
    * `GET /v1/reflection/status/{id}`: Returns the current vote count and quorum status.
    * `POST /v1/reflection/vote`: Endpoint for Auditor agents to submit signed approval/rejection tokens.
    * `GRPC Stream ReflectionEvents`: Real-time notification for agents selected for a quorum.
* **Data Storage/State:**
    * **Quorum Registry**: Ephemeral in-memory store (Redis-backed) tracking active reflection sessions, selected auditors, and received votes.
    * **Expertise Matrix**: Persistent SQLite table mapping agent identities to their validated domains (e.g., "SQL-Security", "Cloud-Networking") to inform auditor selection.

## 5. Alternatives Considered
* **Agent-Local Reflection**: Auditors are selected by the Specialist agent themselves.
    * *Rejected*: High risk of "Collusion Hijacking" where a specialist selects its own "friends" (subagents it spawned) to bypass security.
* **Synchronous Quorums**: Specialist waits until all votes are in before the API returns.
    * *Rejected*: Adds too much latency (MTTC).
    * *Solution*: **Optimistic Reflection** allows the specialist to proceed with "Local Speculative State" while the broker performs the background quorum validation.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):**
    * **Auditor Independence**: The broker mandates that Auditors must have a different lineage than the Specialist to prevent sibling-collusion.
    * **Trace Sanitization**: The Broker scrubs mission-root secrets from reasoning traces before sending them to low-trust Auditors.
* **Observability:**
    * **Reflection Traces**: Log every reflection decision, including Auditor IDs and reasoning summaries, into the immutable audit trail.
    * **Latency Monitoring**: Track "Time to Consensus" to identify slow Auditor agents for de-prioritization.

## 7. Evolutionary Changelog
* **2026-07-09:** Initial Document Creation.
