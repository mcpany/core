# Design Doc: Hierarchical Intent Quota Manager
**Status:** Draft
**Created:** 2026-04-06

## 1. Context and Scope
As agent swarms evolve towards autonomous sub-swarm spawning (e.g., OpenClaw's Ephemeral Sub-Swarms), the risk of "Recursive Resource Storms" increases. A single mission can trigger an exponential growth in sub-agents, leading to uncontrolled token consumption, credit exhaustion, and latency spikes. MCP Any needs a mechanism to enforce hierarchical resource boundaries that sub-swarms cannot exceed.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a cryptographic "Resource Envelope" for agent intents.
    * Enable parent agents to delegate a subset of their own quota to sub-swarms.
    * Provide real-time enforcement of token, credit, and depth limits.
    * Ensure quota lineage is verifiable across the Universal Agent Bus (UAB).
* **Non-Goals:**
    * Managing the underlying billing system for LLM providers (handled by providers).
    * Restricting tool access logic (handled by the Policy Firewall).

## 3. Critical User Journey (CUJ)
* **User Persona:** Multi-Agent Swarm Orchestrator
* **Primary Goal:** Bound the total resource consumption of a complex, multi-threaded reasoning task.
* **The Happy Path (Tasks):**
    1. Orchestrator initializes a primary mission with a 1-million token budget and a recursion depth of 3.
    2. MCP Any generates a root "Quota Token" bound to the Mission Intent.
    3. Primary Agent spawns two specialized sub-swarms for parallel research.
    4. Primary Agent delegates 200k tokens each to the sub-swarms by signing "Sub-Quota Tokens."
    5. Sub-swarms perform tool calls; MCP Any deducts from the sub-quota and parent quota simultaneously.
    6. A sub-swarm attempt to spawn its own sub-sub-swarm at depth 4 is rejected by MCP Any.

## 4. Design & Architecture
* **System Flow:**
    * Quota Allocation: `Root Intent -> Quota Envelope -> Sub-Intent -> Sub-Envelope`.
    * Every UACO/UAB request must carry a `X-MCP-Quota-Token`.
    * Middleware intercepts tool calls, validates the token signature/lineage, and checks remaining balances.
* **APIs / Interfaces:**
    * `POST /v1/quotas/delegate`: Create a signed sub-envelope from an existing quota.
    * `GET /v1/quotas/status`: View remaining balance and lineage for an intent.
* **Data Storage/State:**
    * Quota balances are stored in the Shared KV Store (Blackboard) under a protected, agent-isolated namespace.

## 5. Alternatives Considered
* **Centralized Quota Service**: Rejected due to latency and the need for offline/air-gapped support. Cryptographic delegation allows for decentralized validation.
* **Flat Quota Limits**: Rejected because they don't account for the hierarchical nature of agent swarms, where a parent needs to control its "offspring."

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Quota tokens must be cryptographically bound to the Intent ID and the Parent Identity to prevent "Quota Smuggling."
* **Observability:** Real-time telemetry export to the `Agentic SLA Monitor` and `RL Telemetry Provider`.

## 7. Evolutionary Changelog
* **2026-04-06:** Initial Document Creation.
