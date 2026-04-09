# Design Doc: Federated Reputation Quorum Hub
**Status:** Draft
**Created:** 2026-04-09

## 1. Context and Scope
The "ClawHavoc" malicious skill crisis demonstrated that static tool registries are vulnerable to supply-chain attacks. A single compromised registry can distribute malicious skills to thousands of agents. Furthermore, "Agentic Social Engineering" allows compromised subagents to manipulate their peers within a swarm.

MCP Any needs a decentralized mechanism to verify tool and agent safety. The Federated Reputation Quorum Hub (FRQH) moves safety attestation from a single authority to a collective consensus of independent security nodes within the Universal Agent Bus (UAB) mesh.

## 2. Goals & Non-Goals
* **Goals:**
    * Aggregate reputation signals (behavioral profiles, safety scores) from multiple independent nodes.
    * Implement a quorum-based voting mechanism to determine tool "Trust Tiers."
    * Provide a real-time "Consensus Receipt" for agents to verify tool safety before execution.
    * Support "Behavioral Drifting" detection where a previously safe tool is downgraded based on new collective evidence.
* **Non-Goals:**
    * Acting as a primary tool registry (it only provides reputation layers for existing registries).
    * Executing the tools themselves (scoped to reputation and attestation).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Verify if a popular community-contributed skill is safe to use in a production environment.
* **The Happy Path (Tasks):**
    1. The orchestrator identifies a new skill and queries FRQH for its reputation.
    2. FRQH broadcasts a reputation request to the UAB mesh.
    3. Multiple independent security nodes return signed behavioral reports and safety scores.
    4. FRQH aggregates the signals and identifies that 80% of nodes attest to its safety.
    5. FRQH generates a "Trust Tier 1" consensus receipt.
    6. The orchestrator allows the tool call, confident in the collective verification.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    sequenceDiagram
        participant Orchestrator
        participant FRQH
        participant UAB_Nodes as Peer Security Nodes
        Orchestrator->>FRQH: Request Reputation(Skill_ID)
        FRQH->>UAB_Nodes: Broadcast Reputation Query
        UAB_Nodes-->>FRQH: Signed Behavioral Reports
        FRQH->>FRQH: Aggregate Consensus (Quorum)
        FRQH-->>Orchestrator: Signed Consensus Receipt
    ```
* **APIs / Interfaces:**
    * `GetConsensusReputation(tool_id string) (ConsensusReceipt, error)`: Retrieves the current collective reputation of a tool.
    * `SubmitBehavioralSignal(signal BehavioralSignal) error`: Allows nodes to contribute new behavioral evidence to the hub.
* **Data Storage/State:** Uses a distributed ledger or a highly available gossip protocol to synchronize reputation states across the mesh.

## 5. Alternatives Considered
* **Centralized Security Authority:** Rejected because it creates a single point of failure and a high-value target for attackers.
* **Local-Only Behavioral Profiling:** Effective but slow; it doesn't benefit from the "Herd Immunity" of a mesh.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Nodes must provide hardware-attested identities to prevent "Sybil Attacks" where an attacker creates multiple fake nodes to manipulate consensus.
* **Observability:** The UI provides a "Federated Reputation Monitor" to visualize quorum progress and trust transitions.

## 7. Evolutionary Changelog
* **2026-04-09:** Initial Document Creation.
