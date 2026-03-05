# Design Doc: Swarm Consensus Middleware
**Status:** Draft
**Created:** 2026-03-05

## 1. Context and Scope
With the rise of decentralized agent swarms like OpenClaw, tool discovery and schema synchronization are moving towards "gossip" protocols. While this increases resilience, it introduces "Consistency Gaps" where different agents in a swarm may see different versions of a tool or even malicious schemas injected via "Gossip Injection." MCP Any needs to provide a "Source of Truth" or "Consensus Provider" that ensures all agents in a swarm are operating on synchronized, verified tool sets.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide a centralized "Consensus Registry" that validates tool schemas before they are propagated to the swarm.
    * Implement a "Consistency Guard" that rejects tool execution requests if the agent is using a stale or unverified schema version.
    * Mitigate "Gossip Injection" by requiring cryptographic signatures for all tool updates.
* **Non-Goals:**
    * Replacing the decentralized gossip protocol entirely.
    * Acting as the primary communication bus for the swarm (it only handles tool/schema consensus).

## 3. Critical User Journey (CUJ)
* **User Persona:** Swarm Security Architect
* **Primary Goal:** Ensure that a swarm of 50 decentralized agents all use the exact same version of a "Financial Reporting" tool and prevent rogue agents from broadcasting malicious updates.
* **The Happy Path (Tasks):**
    1. The Architect registers the "Financial Reporting" tool with MCP Any and signs its schema.
    2. Agents in the swarm poll MCP Any or receive updates via the Consensus Guard.
    3. When an agent attempts to call the tool, MCP Any verifies the schema hash against the Consensus Registry.
    4. If a rogue agent tries to broadcast a "v2" with a backdoor, MCP Any detects the signature mismatch and alerts the swarm.

## 4. Design & Architecture
* **System Flow:**
    `Agent (Swarm) <-> Gossip Protocol <-> MCP Any (Consensus Guard) <-> Tool Registry`
* **APIs / Interfaces:**
    * `POST /consensus/verify`: Validates a tool schema and returns a signed "Lease."
    * `GET /consensus/sync`: Returns the current global state of all verified tools in the swarm.
    * `POST /consensus/attest`: Submits a cryptographic signature for a new tool version.
* **Data Storage/State:**
    * Uses a Merkle Tree structure in the `Shared KV Store` to allow for efficient consistency checks across large swarms.

## 5. Alternatives Considered
* **Pure Decentralized Consensus (Paxos/Raft)**: Rejected due to high latency and complexity for agent frameworks to implement natively.
* **Static Configs**: Rejected because swarms are dynamic and tools are often added/removed at runtime.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: All consensus updates require "Provenance-First" attestation. Tools without a valid chain of trust are quarantined.
* **Observability**: The UI will show a "Swarm Consistency Heatmap," highlighting any agents that are out of sync or using unverified tools.

## 7. Evolutionary Changelog
* **2026-03-05**: Initial Document Creation.
