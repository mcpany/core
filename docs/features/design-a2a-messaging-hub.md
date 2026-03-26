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

### Update: 2026-04-18 - Foundation Governance & Continuous Attestation
**Context:** The transition of OpenClaw to an independent foundation and the maturation of "Sandbox Persistence Proofs" require the Hub to support institutionalized governance and continuous security signals.
**Architecture Adjustment:**
* **Foundation Governance Bridge:** Added support for the OpenClaw Foundation's neutral governance protocols to the delegation logic in Section 4.
* **Unified Persistence Broker Integration:** The Hub now mandates a valid "Sandbox Persistence Proof" from the RIM before authorizing high-sensitivity A2A task handoffs.
**Security Impact:** Ensures inter-agent collaboration is compliant with foundation-neutral mandates and remains secure against post-boot sandbox exploits.

### Update: 2026-04-20 - A2A Safety Proofs & Coercion Defense
**Context:** Today's research into the OpenClaw "M2M Loop" crisis and the successful transition of the A2A protocol to the Linux Foundation highlights a new threat: "Inter-Agent Coercion," where a compromised subagent attempts to manipulate its parent via a task proposal.
**Architecture Adjustment:**
* **Safety Proof Mandatory Validation:** Updating the `/v1/a2a/propose` logic in Section 4 to mandate a "Safety Proof" for all task proposals. This proof must include a cryptographically signed justification and a reputation-bound capability claim.
* **Coercion Detection Middleware:** Introducing an interception layer that scans task proposals for imperative instructions targeting the parent agent's reasoning engine (e.g., "forget previous instructions").
**Security Impact:** Neutralizes the "ClawHavoc" style coercion vector by ensuring all inter-agent task delegations are authenticated, scoped, and semantically sanitized.

### Update: 2026-05-17 - Authenticated Agent Card Discovery
**Context:** Gemini CLI v0.33.0 has introduced HTTP authentication for A2A remote agents and "Authenticated Agent Card Discovery" to prevent unauthorized capability claims.
**Architecture Adjustment:**
* Implementing **Auth-Before-Discovery** in Section 4.
* The Hub will now require a valid mission-bound identity token before exposing "Agent Cards" or capability metadata to a requesting peer.
* Integrating the `TeammateTool` protocol to support Claude-style hierarchical discovery across heterogeneous frameworks.
**Security Impact:** Prevents "Shadow Capability" mapping by malicious subagents and ensures that only authorized teammates can discover and spawn specialized agents within the bus.

### Update: 2026-06-01 - Machine-Speed Quarantining & Mandatory Discovery Auth
**Context:** The 2026 Armis report confirm that "Agentic Swarms" can weaponize exploits in seconds, requiring the A2A Hub to move from passive routing to active interdiction.
**Architecture Adjustment:**
*   Integrating **MSSQ (Machine-Speed Swarm Quarantine)** triggers into the `/v1/a2a/mailbox` and delivery logic. The Hub will now check the high-speed bitset for quarantine status before delivering any task proposal or message.
*   Enforcing **Mandatory Discovery Auth** for all UAB-connected peers, aligning with the Gemini CLI v0.33.0 baseline. Capability metadata is now cryptographically masked until a verified mission-bound handshake is completed.
**Security Impact:** Neutralizes "Hivenet" propagation by cutting off inter-agent communication channels in sub-milliseconds and ensures "Zero-Visibility" for unauthenticated probes.
<<<<<<< HEAD

### Update: 2026-03-23 - A2A Authentication Proxy & Sandbox Identity
**Context:** Gemini CLI v0.34.0 has standardized HTTP authentication for remote A2A agents and authenticated agent card discovery. Simultaneously, the rise of RL-driven swarms demands hardware-bound environment attestation.
**Architecture Adjustment:**
* **A2A Authentication Proxy:** Introducing a mandatory authentication gate for the `/v1/a2a/propose` and `/v1/a2a/mailbox` endpoints in Section 4. The hub will now validate bearer tokens against a verified peer registry.
* **Execution Identity Integration:** Extending the A2A Task Card in Section 4 to include a `sandbox_attestation` field, allowing agents running in gVisor to prove their environment integrity.
**Security Impact:** Eliminates unauthenticated capability probes and provides a foundation for zero-trust inter-agent delegation in heterogeneous meshes.
=======
>>>>>>> 4f039895e (⚡ Bolt: Render Optimization for System Status Banner (#6544))
