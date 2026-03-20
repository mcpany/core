# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

# Design Doc: Capability Beacon Firewall (CBF)
**Status:** Draft
**Created:** 2026-05-15

## 1. Context and Scope
With the rise of push-based tool discovery, malicious or malfunctioning MCP servers can flood the agent ecosystem with unauthenticated "Capability Beacons". This "Capability Beacon Storm" acts as a Denial of Service (DoS) attack on the agent's reasoning engine, which becomes overwhelmed trying to index and prioritize thousands of new tool claims per second. The Capability Beacon Firewall (CBF) is needed to provide a high-performance, rate-limited ingestion layer that protects MCP Any and its connected agents from discovery-phase exhaustion.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement sub-millisecond rate-limiting for reactive discovery beacons (UDP/HTTP).
    * Provide a quarantine mechanism for beacons from un-attested or high-frequency sources.
    * Support cryptographic signature verification for "Trusted Beacons".
    * Integrate with the Federated Discovery Quorum (FDQ) for reputation-based filtering.
* **Non-Goals:**
    * The CBF will NOT perform deep semantic analysis of the tool schema (handled by the Metadata Sanitizer).
    * It will NOT block authenticated discovery requests from verified local frameworks.

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Protect the swarm's reasoning capacity during a machine-speed discovery flood.
* **The Happy Path (Tasks):**
    1. A malicious local tool begins broadcasting 5,000 UDP Capability Beacons per second.
    2. The CBF intercepts the beacons at the transport layer.
    3. The CBF identifies that the source exceeds the "Discovery-Phase Burst Threshold".
    4. The CBF automatically quarantines the source and drops subsequent beacons from that IP/Origin.
    5. The orchestrator receives a single "Discovery Throttled" alert instead of 5,000 new tool notifications.
    6. The swarm continues its primary mission without reasoning latency degradation.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        Beacons[Discovery Beacons] --> Receiver[Beacon Receiver]
        Receiver --> RL[Rate Limiter & Bursts Controller]
        RL -->|Throttled| Q[Quarantine Store]
        RL -->|Allowed| V[Signature Validator]
        V -->|Verified| Registry[Service Registry]
        V -->|Unverified| FDQ[Federated Discovery Quorum]
    ```
* **APIs / Interfaces:**
    * `discovery.beacon_threshold`: Configuration for maximum beacons per second per source.
    * `discovery.quarantine_ttl`: Duration to ignore a throttled source.
    * `GetQuarantinedSources()`: Admin API to review blocked discovery origins.
* **Data Storage/State:**
    * Ephemeral, in-memory Bloom filter for high-speed source tracking.
    * SQLite backend for persistent quarantine logs and reputation history.

## 5. Alternatives Considered
* **Agent-Side Filtering:** Rejected because it still requires the reasoning engine to ingest and "ignore" the beacons, consuming context and compute.
* **Static Allow-Listing:** Rejected as too rigid for dynamic, local-first development environments where new tools are frequently added.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The CBF acts as the "Zero-Trust Gatekeeper" for discovery, ensuring no tool enters the registry without passing rate and signature checks.
* **Observability:** Prometheus metrics for `beacons_received_total`, `beacons_throttled_total`, and `quarantine_active_sources`.

## 7. Evolutionary Changelog
* **2026-05-15:** Initial Document Creation.
