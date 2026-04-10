# Design Doc: Consensus-Driven Resource Rebalancing (CDRR) Manager
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As AI agents evolve from linear sessions into high-density horizontal meshes (e.g., Claude Code Agent Teams), the limitation of static resource allocation becomes a primary performance bottleneck. "Cognitive Stall" occurs when a critical specialist subagent exhausts its hardware-attested reasoning or token budget while parallel teammates maintain significant surplus. Current infrastructure lacks the liquidity to rebalance these resources without global mission-root intervention, leading to coordination deadlocks and latency spikes.

MCP Any needs to solve this by introducing the CDRR Manager. This system acts as an authoritative, mesh-resident auctioneer that allows subagents to dynamically trade or claim surplus resources. By leveraging hardware-attested "Surplus Claims" and "Deficit Signals," MCP Any can ensure that the mesh operates at peak efficiency while maintaining strict mission-root economic sovereignty.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a real-time, non-blocking auction bus for reasoning turns and token budgets.
    * Mandate hardware-attested (TPM/Secure Enclave) signatures for all resource rebalancing claims.
    * Ensure mission-root budget continuity across heterogeneous framework boundaries (OpenClaw, Claude, AutoGen).
    * Provide sub-10ms resolution for budget re-allocation events to prevent reasoning latency.
* **Non-Goals:**
    * This system WILL NOT allow agents to exceed the total aggregate budget defined by the human user for the primary mission-root.
    * It WILL NOT be used for inter-mission resource sharing; budgets are strictly mission-bound.

## 3. Critical User Journey (CUJ)
* **User Persona:** High-Density Agent Team Orchestrator
* **Primary Goal:** Dynamically rebalance budgets between a "Security Auditor" agent (budget surplus) and a "Code Generator" agent (budget exhausted) to complete a complex PR without stalling.
* **The Happy Path (Tasks):**
    1. The Code Generator agent detects it has reached 90% of its reasoning turn limit.
    2. It broadcasts a hardware-attested "Deficit Signal" to the CDRR Manager.
    3. The Security Auditor agent, having completed its task early, broadcasts a "Surplus Claim."
    4. The CDRR Manager matches the signals and initiates a "Rebalance Auction."
    5. A multi-agent quorum (including an Independent Auditor) verifies the mission-root alignment.
    6. The CDRR Manager atomically updates the hardware-locked budgets for both agents.
    7. The Code Generator resumes reasoning without a cold-boot or manual re-authorization.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    sequenceDiagram
        participant SG as Subagent (Deficit)
        participant CM as CDRR Manager
        participant MR as Mission Root Authority
        participant SA as Subagent (Surplus)

        SG->>CM: Attested Deficit Signal (Reasoning: 5 turns)
        SA->>CM: Attested Surplus Claim (Reasoning: 20 turns)
        CM->>MR: Validate Rebalance against Mission Policy
        MR-->>CM: Policy Approval (Signed)
        CM->>CM: Execute Atomic Shard Rebalance
        CM-->>SG: Updated Budget Token (Hardware-bound)
        CM-->>SA: Budget Revocation Receipt
    ```
* **APIs / Interfaces:**
    * `POST /v1/mesh/resource/signal`: Submit deficit or surplus claims.
    * `GET /v1/mesh/resource/auction/status`: Monitor active rebalancing events.
    * `x-mcp-resource-attestation`: New header for transporting hardware-bound budget tokens.
* **Data Storage/State:**
    * State is managed via **Asynchronous Mailbox Shards (AMS)**.
    * Budget allocations are stored in kernel-bound, hardware-locked memory buffers to prevent tampering.

## 5. Alternatives Considered
* **Static Round-Robin Rebalancing**: Rejected due to high "Coordination Tax" and inability to handle non-deterministic reasoning spikes.
* **Centralized Mission-Root Control**: Rejected because it creates a single point of failure and adds 200ms+ latency to every budget check.
* **OpenClaw-Native CDRR**: Rejected as a primary solution because it is framework-specific. MCP Any must provide the universal bridge for heterogeneous swarms.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):**
    * Every signal must be signed by the agent's mesh-resident identity.
    * Rebalancing is only authorized if the destination agent can prove its lineage back to the *same* mission-root as the donor.
* **Observability:**
    * Real-time "Mesh Liquidity" metrics will be exported to the **Dynamic Resource Budgeting Monitor**.
    * Audit logs will include the complete hardware-attested chain of every resource transfer.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
