# Design Doc: Reason-Graph Integrity (RGI) Provider
**Status:** Draft
**Created:** [2026-06-18]

## 1. Context and Scope
As agent swarms become more complex and autonomous, they rely on hierarchical reasoning paths to solve multi-step problems. Today's market sync revealed the "Reason-Graph Collision" (RGC) exploit, where a malicious subagent can inject reasoning nodes that create structural cycles or cognitive deadlocks in the parent's reasoning process.

The Reason-Graph Integrity (RGI) Provider is a core security service that performs hardware-attested structural analysis of an agent's reason-graph. It ensures that all reasoning paths remain acyclic and consistent with the hardware-attested Mission-Root anchors, neutralizing structural cognitive attacks.

## 2. Goals & Non-Goals
* **Goals:**
    * Perform real-time structural analysis of agent reason-graphs.
    * Detect and block reasoning nodes that create cycles (deadlocks).
    * Validate reasoning-path lineage against Mission-Root intent anchors.
    * Provide hardware-attested (TPM) signatures for verified reason-graph segments.
* **Non-Goals:**
    * Validating the semantic truth of an agent's reasoning (handled by AID Hub).
    * Managing token budgets (handled by RBF).
    * Providing long-term memory (handled by Blackboard).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Prevent a specialized subagent from stalling the mission via structural reasoning collisions.
* **The Happy Path (Tasks):**
    1. Parent Agent spawns a specialist subagent to perform code analysis.
    2. Subagent attempts to inject a reasoning loop that claims task A depends on B, and B depends on A.
    3. RGI Provider intercepts the graph update and performs cycle detection.
    4. RGI identifies the structural collision and rejects the reasoning node.
    5. RGI alerts the Parent Agent and prunes the compromised subagent's path.
    6. Mission continues using the verified, acyclic reasoning path.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Subagent Reasoning Node] --> B[RGI Provider]
        B --> C[Structural Parser]
        C --> D[Cycle Detection Engine]
        D --> E{Collision Detected?}
        E -- Yes --> F[Block Node & Alert Parent]
        E -- No --> G[Attest & Append to Reason-Graph]
        H[Mission-Root Anchor] --> C
    ```
* **APIs / Interfaces:**
    * `rgi.ValidatePath(node, graph) -> Result`: Validates a new reasoning node against the current graph.
    * `rgi.GetAttestedGraph(sessionID) -> SignedGraph`: Returns a hardware-signed snapshot of the reason-graph.
* **Data Storage/State:**
    * **Active Reason-Graph:** An in-memory Directed Acyclic Graph (DAG) representing the current mission's reasoning lineage.

## 5. Alternatives Considered
* **Time-based Deadlock Detection:** Rejected because it only reacts after the stall occurs. RGI provides proactive structural prevention.
* **Semantic Verification Only:** Rejected because structural exploits can be semantically plausible but architecturally fatal.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** RGI utilizes hardware-bound (TPM) primitives for all graph attestation.
* **Observability:** Graph collisions are logged to the "Reason-Graph Integrity Monitor" for forensic analysis.

## 7. Evolutionary Changelog
* **[2026-06-18]:** Initial Document Creation. Addressing Reason-Graph Collision (RGC) vulnerabilities.
