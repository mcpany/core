# Design Doc: DCA Negotiation Guard
**Status:** Draft
**Created:** 2026-04-05

## 1. Context and Scope
As agent swarms grow in complexity, the "Distributed Capability Auction" (DCA) has introduced a new failure mode: **Negotiation Exhaustion**. High-velocity swarms spend more compute cycles on task bidding than on execution, leading to "Negotiation Storms."

The DCA Negotiation Guard is a hardware-accelerated (HAN) broker in MCP Any that provides a trusted, low-latency environment for subagent bidding and resource allocation.

## 2. Goals & Non-Goals
* **Goals:**
    * Reduce negotiation latency in multi-agent swarms by offloading bidding logic to a dedicated broker.
    * Prevent "Negotiation Storms" by implementing rate limits and bid-priority queues.
    * Support Hardware-Accelerated Negotiation (HAN) using TPM/Secure Enclave for bid attestation.
* **Non-Goals:**
    * Defining the bidding strategies for individual agents.
    * Replacing the UACO protocol (it extends it).

## 3. Critical User Journey (CUJ)
* **User Persona:** Multi-Agent Swarm Orchestrator
* **Primary Goal:** Coordinate 20 specialized subagents without the orchestration layer becoming a bottleneck.
* **The Happy Path (Tasks):**
    1. Parent Agent issues a UACO Task Card via MCP Any.
    2. Subagents submit cryptographically signed bids to the DCA Negotiation Guard.
    3. The Guard uses hardware acceleration to verify signatures and rank bids in sub-millisecond time.
    4. The winning agent is notified and granted the "Capability Lease."
    5. The swarm proceeds without "Negotiation Exhaustion" lag.

## 4. Design & Architecture
* **System Flow:**
    * **HAN Broker**: Core component that interacts with hardware security modules for fast-path attestation.
    * **Priority Queue**: Manages bids based on agent reputation and task urgency.
    * **Lease Manager**: Tracks the duration and scope of granted capabilities.
* **APIs / Interfaces:**
    * `POST /uaco/bid`: Submit a task bid.
    * `GET /uaco/auction/{task_id}`: Monitor auction status.
* **Data Storage/State:**
    * Ephemeral in-memory state for active auctions.

## 5. Alternatives Considered
* **Software-Only Brokering**: Rejected as it cannot meet the sub-millisecond latency requirements for large-scale swarms.
* **Centralized Orchestrator**: Rejected to maintain the decentralized nature of UACO-compliant swarms.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All bids must be hardware-attested to prevent bid spoofing or "Task Shadowing."
* **Observability:** Telemetry on auction duration, bid density, and negotiation overhead.

## 7. Evolutionary Changelog
* **2026-04-05:** Initial Document Creation.

### Update: 2026-04-06 - Support for Speculative Auction Brokering (SAB)
**Context:** Experimental "Intent Probability" bidding requires a broker that can handle speculative tasks.
**Architecture Adjustment:**
* Extending the DCA Guard to support the **SAB Protocol**.
* The Guard will now manage "Shadow Auctions" for speculative branches, allowing agents to bid on probable tasks before they are finalized.
**Security Impact:** Ensures that speculative task delegation maintains the same Zero-Trust attestation as deterministic flows.

### Update: 2026-04-04 - Integration with MSSI for Negotiation Storm Mitigation
**Context:** Today's market sync revealed that coordinated Hivenet attacks can spoof bidding density to trigger "Negotiation Exhaustion."
**Architecture Adjustment:**
* Integrating the **Machine-Speed Swarm Interdiction (MSSI)** hub directly into the HAN Broker loop.
* The Guard now performs behavioral anomaly scoring on bid patterns. If bid entropy suggests a "Negotiation Storm" attack, the MSSI trigger will automatically revoke capability discovery for the suspect subagent branch.
**Security Impact:** Prevents Agentic DoS attacks designed to freeze the swarm's decision-making logic.
