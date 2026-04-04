# Design Doc: Hardware-Accelerated Negotiation (HAN)
**Status:** Draft
**Created:** 2026-04-04

## 1. Context and Scope
As AI agent swarms grow in complexity and depth, the coordination overhead—specifically task bidding via protocols like Distributed Capability Auction (DCA)—has become a significant bottleneck. "Negotiation Exhaustion" occurs when agents spend more resources on coordination than on task execution, leading to "Negotiation Storms" that freeze the swarm.

MCP Any needs to solve this by providing a high-speed, authoritative arbitration layer that offloads the bidding logic from the LLM reasoning loop to hardware-backed system infrastructure.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide a low-latency bus for multi-agent task auctions.
    * Utilize hardware-backed (TPM/SEP) arbitration to ensure non-repudiable and secure bidding.
    * Implement real-time "Negotiation Storm" detection and mitigation.
    * Support standardized bidding formats across UACO and A2A.
* **Non-Goals:**
    * Replacing the underlying LLM's decision-making logic for complex strategy.
    * Managing the execution of the tasks themselves (handled by the tool gateway).

## 3. Critical User Journey (CUJ)
* **User Persona:** High-Density Swarm Orchestrator
* **Primary Goal:** Successfully delegate 50+ concurrent tasks to 10+ specialist agents without triggering a coordination stall.
* **The Happy Path (Tasks):**
    1. Parent agent posts a task card to the MCP Any UACO bus.
    2. Multiple specialist subagents submit bids (Capability Cards) to the HAN Broker.
    3. The HAN Broker validates bid integrity via hardware attestation.
    4. The HAN Broker applies mission-aligned "Fairness Policies" to select a winner in <5ms.
    5. The parent agent receives the winning bid and authorizes the handoff.

## 4. Design & Architecture
* **System Flow:**
    [Agent] -> (Task Bid) -> [HAN Broker (System Space)] -> (Arbitration) -> [Winning Agent]
    The HAN Broker lives in the system space, utilizing a dedicated worker thread and hardware enclave hooks.
* **APIs / Interfaces:**
    * `POST /v1/negotiation/auction`: Initiate a new task auction.
    * `POST /v1/negotiation/bid`: Submit a hardware-signed bid for an active auction.
* **Data Storage/State:**
    * In-memory priority queues for active auctions.
    * Bloom filters for rapid "Capability Card" verification.

## 5. Alternatives Considered
* **Pure-LLM Negotiation**: Too slow (seconds vs milliseconds) and prone to hallucination.
* **Software-Only Arbiter**: Vulnerable to "Bid Hijacking" if the coordination process is compromised; lacks the non-repudiation of hardware attestation.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All bids must be cryptographically bound to the agent's hardware identity.
* **Observability:** Metrics for "MTTN" (Mean Time to Negotiate) and "Auction Depth" are exported to the telemetry sink.

## 7. Evolutionary Changelog
* **2026-04-04:** Initial Document Creation.
