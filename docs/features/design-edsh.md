# Design Doc: Event-Driven Swarm Hub (EDSH)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As AI agent swarms move from linear, supervisor-led sequences to horizontal, peer-to-peer teammate meshes (e.g., OpenClaw v3.6, Claude Code Agent Teams), the "Mailbox" pattern is hitting performance and scalability ceilings. Agents currently rely on polling shared task lists or being explicitly "called" by a supervisor, which introduces MTTC (Mean Time to Coordinate) latency and creates a central bottleneck.

EDSH (Event-Driven Swarm Hub) evolves the Universal Agent Bus into an asynchronous signaling layer. By introducing native pub/sub semantics, agents can broadcast state shifts, tool outputs, and task completions to any interested teammate instantly, enabling reactive, high-speed coordination without lock-contention.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a low-latency pub/sub broker for agent-to-agent signaling.
    * Support "Topic-Based" discovery where agents can subscribe to specific task types or state updates.
    * Provide hardware-attested event signatures to ensure signal integrity.
    * Integrate with the existing AMS (Asynchronous Mailbox Sharding) for hybrid coordination.
* **Non-Goals:**
    * Replacing the Blackboard (Shared KV Store) for persistent state.
    * Managing the LLM reasoning loop itself (this is a coordination middleware).
    * Providing long-term event archiving (events are ephemeral by default).

## 3. Critical User Journey (CUJ)
* **User Persona:** Multi-Agent DevSecOps Swarm
* **Primary Goal:** Coordinate a parallel refactor where 5 agents implement features while a 6th agent (QA) reacts to "Implementation Complete" signals to run tests.
* **The Happy Path (Tasks):**
    1. The Team Lead agent publishes a `task.available` event for a refactor sub-task.
    2. A Specialist agent receives the event via its subscription and "Claims" the task.
    3. Upon finishing, the Specialist publishes a `feature.implemented` event with a Git commit hash.
    4. The QA agent, subscribed to `feature.implemented`, automatically receives the event and triggers the test suite.
    5. All coordination happens without the Team Lead having to poll the Blackboard.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        AgentA[Agent A] -->|Publish Event| EDSH
        EDSH -->|Route| Broker[Topic Router]
        Broker -->|Hardware Attest| HSM[TPM/Secure Enclave]
        HSM -->|Attested Signal| AgentB[Agent B]
        HSM -->|Attested Signal| AgentC[Agent C]
    ```
* **APIs / Interfaces:**
    * `POST /v1/events/publish`: `{ "topic": string, "payload": object, "origin_signature": string }`
    * `GET /v1/events/subscribe`: SSE or WebSocket stream filtered by topic pattern.
    * `POST /v1/events/topics`: Register a new capability-bound topic.
* **Data Storage/State:**
    * In-memory event ring buffer for high-speed routing.
    * Hardware-attested event log for auditability (ephemeral).

## 5. Alternatives Considered
* **Blackboard Polling**: Rejected due to high latency and O(N^2) complexity as swarm size increases.
* **Direct WebSocket P2P**: Rejected because it requires agents to know each other's addresses, violating the "Universal Bus" abstraction and Zero-Trust discovery.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All topics are capability-gated. An agent can only publish to topics it has an "Intent-Bound" token for.
* **Observability:** Real-time event tracing in the "Agent Chain Tracer" UI, showing the flow of signals between teammates.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
