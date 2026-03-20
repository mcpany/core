# Design Doc: Foundation Governance Adapter
**Status:** Draft
**Created:** 2026-04-18

## 1. Context and Scope
With the transition of the OpenClaw project to an independent foundation sponsored by OpenAI, there is an urgent need for agentic infrastructure to support standardized, foundation-neutral governance protocols. Agents from disparate frameworks increasingly need a common translation layer to ensure that task delegations and state handoffs comply with the Foundation's transparency and security mandates.

The Foundation Governance Adapter (FGA) acts as this translation layer within MCP Any, bridging local agent intents with the Foundation's global governance policies.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a bridge for the OpenClaw Foundation's neutral governance protocols.
    * Provide a standardized interface for cross-framework agent coordination.
    * Ensure all inter-agent delegations are auditable and compliant with Foundation-mandated transparency rules.
    * Map framework-specific identity and intent to foundation-neutral governance tokens.
* **Non-Goals:**
    * Replacing existing framework-specific security (e.g., OpenClaw's internal PoL).
    * Providing a marketplace for skills (this is handled by the Verified Skill Registry).

## 3. Critical User Journey (CUJ)
* **User Persona:** Multi-Framework Swarm Architect
* **Primary Goal:** Securely delegate a high-sensitivity task from an OpenClaw subagent to an AutoGen-based specialist while ensuring the handoff is compliant with the OpenClaw Foundation's audit standards.
* **The Happy Path (Tasks):**
    1. The OpenClaw agent initiates a task delegation request via UAB.
    2. MCP Any intercepts the request and identifies the requirement for Foundation-level governance.
    3. The FGA translates the OpenClaw intent into a Foundation-Neutral Governance Token.
    4. The token is validated against the Foundation's local policy mirror.
    5. The task is delegated to the AutoGen specialist with the governance token attached.
    6. An immutable audit record is generated and stored in the Lineage-Aware Audit Log.

## 4. Design & Architecture
* **System Flow:**
    `Agent Request -> UAB Adapter -> Foundation Governance Adapter (FGA) -> Policy Mirror -> Target Agent`
* **APIs / Interfaces:**
    * `/v1/governance/attest`: Generates a neutral governance token from framework-specific intents.
    * `/v1/governance/verify`: Validates a governance token against local and remote policy manifests.
* **Data Storage/State:**
    * State is managed via a dedicated "Governance State" table in the internal SQLite Blackboard, tracking token lineage and audit hashes.

## 5. Alternatives Considered
* **Direct Integration in Frameworks:** Rejected because it forces framework authors to implement complex governance logic, creating fragmentation. MCP Any is the natural neutral ground for this translation.
* **Cloud-Only Governance:** Rejected to maintain the "Local-First" priority of MCP Any and ensure privacy.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All governance tokens are cryptographically bound to the initiating agent's identity and the current mission intent.
* **Observability:** Integrated with the Unified RL Feedback Telemetry Bridge for monitoring governance compliance and latency.

## 7. Evolutionary Changelog
* **2026-04-18:** Initial Document Creation.
