# Design Doc: A2A Messaging Hub
**Status:** Draft
**Created:** 2026-04-12

## 1. Context and Scope
With the transition of the Agent2Agent (A2A) protocol to the Linux Foundation, it has become the industry standard for inter-agent communication. MCP Any must evolve from a simple protocol bridge to a native A2A Messaging Hub. This hub will manage the discovery, negotiation, and secure delegation of tasks between disparate agent frameworks (e.g., OpenClaw, AutoGen) while enforcing local Zero-Trust security policies.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide a native, high-performance implementation of the A2A messaging protocol.
    * Act as a security broker that validates A2A task cards against local capability tokens.
    * Maintain a persistent, stateful "mailbox" for asynchronous agent coordination.
    * Integrate with the Shared KV Store (Blackboard) for cross-framework state persistence.
* **Non-Goals:**
    * Replacing existing agent frameworks (e.g., we do not provide a reasoning engine).
    * Providing a public, unauthenticated relay for arbitrary agent traffic.

## 3. Critical User Journey (CUJ)
* **User Persona:** Multi-Agent Swarm Orchestrator
* **Primary Goal:** Securely delegate a file-writing task from a cloud-based OpenClaw agent to a local AutoGen subagent without exposing host environment variables.
* **The Happy Path (Tasks):**
    1. The OpenClaw agent sends an A2A "Task Proposal" to MCP Any.
    2. MCP Any validates the proposal's cryptographic identity and "Intent Scope."
    3. MCP Any checks the proposal against the "Settings Injection Guard" to ensure no malicious hooks are involved.
    4. MCP Any routes the task to the local AutoGen subagent via an authenticated A2A task card.
    5. The subagent executes the task and returns a "Task Completion" card through the hub.
    6. MCP Any records the transaction in the immutable audit trail and updates the Blackboard.

## 4. Design & Architecture
* **System Flow:**
    `[Agent A (OpenClaw)] -> (A2A Protocol) -> [MCP Any A2A Hub] -> (Policy Firewall) -> [Agent B (AutoGen)]`
* **APIs / Interfaces:**
    * `/v1/a2a/propose`: Endpoint for submitting task proposals.
    * `/v1/a2a/mailbox`: SSE/WebSocket interface for real-time task delivery.
    * `A2A Task Card (JSON-RPC/gRPC)`: Standardized carrier for task metadata and attestation.
* **Data Storage/State:**
    * Uses the embedded SQLite Blackboard for task state and "Intent Chain" storage.
    * Hardware-bound session tokens for inter-agent authentication.

## 5. Alternatives Considered
* **Pseudo-MCP Bridge (Existing):** Rejected for long-term use as it lacks native support for A2A's negotiation and bidding phases, leading to "Intent Ghosting."
* **Direct A2A Integration in Agents:** Rejected as it forces every agent framework to implement its own complex security and discovery logic, leading to inconsistent enforcement.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All inter-agent messages must carry a Proof-of-Intent (PoI) signed by the mission root. The hub enforces Recursive Intent Delegation (RID) limits.
* **Observability:** Integrated with the "Agent Chain Tracer" to provide a visual timeline of all inter-agent handoffs and task states.

## 7. Evolutionary Changelog
* **2026-04-12:** Initial Document Creation.

### Update: 2026-04-13 - Aligning with Linux Foundation Open Governance
**Context:** The A2A protocol has finalized its transition to the Linux Foundation, establishing a vendor-neutral standard for inter-agent coordination.
**Architecture Adjustment:**
* The hub now implements the finalized LF security manifest for task proposals.
* Introducing the "A2A Security Posture Broker" role to validate cross-framework delegations against open-standard policies.
**Security Impact:** Ensures neutral, standard-compliant enforcement of Zero-Trust policies across disparate agent frameworks.

### Update: 2026-04-14 - Verifiable Task Delegation & Context Sidecars
**Context:** Today's research into the OpenClaw v2026.3.7 stabilization and the "44% Manual Review" bottleneck in agent swarms demands a move toward automated, verifiable trust and deeper framework integration.
**Architecture Adjustment:**
* **Delegation Attestation:** Integrating a new "Safety Proof" generator into Section 4. The hub will now evaluate task proposals against historical reputation and policy compliance before surfacing them.
* **Context Sidecar Adapter:** Introducing a "Sidecar" pattern in Section 4 to synchronize state with external frameworks (like OpenClaw's `ContextEngine`) via their native plugin APIs.
**Security Impact:** Reduces manual oversight requirements by providing a verifiable trust signal for autonomous delegations and ensures context integrity across framework boundaries.

### Update: 2026-04-16 - Mission-Anchored Context & PID-Bound Port Integrity
**Context:** Today's research reveals a new "Ephemeral Port Hijacking" exploit (CVE-2026-28501) and the emergence of "Contextual Anchors" for swarm stability.
**Architecture Adjustment:**
* **PID-Bound Port Attestation:** Section 4 updated to include kernel-level PID binding for the Hub's local listeners.
* **Contextual Anchor Middleware:** Introducing an "Anchor Middleware" in the A2A routing pipeline to pin mission intent to subagent prompts.
**Security Impact:** Neutralizes port hijacking race conditions and prevents "Intent Ghosting" in parallel agent swarms.
