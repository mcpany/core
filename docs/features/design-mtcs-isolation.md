# Design Doc: Multi-Tenant Cognitive Shard (MTCS) Isolation
**Status:** Draft
**Created:** 2026-04-07

## 1. Context and Scope
As MCP Any evolves into a core infrastructure layer for enterprise swarms, the need to host multiple "Agent Tenants" (e.g., different departments or projects) on shared infrastructure has become critical. The emergence of OpenClaw's Agentic Multi-Tenant Shards (AMTS) highlights the risk of "Context Smearing," where sensitive mission data or behavioral guardrails from one tenant might leak into another's reasoning window. MCP Any requires a secure hypervisor-like layer to provide absolute cognitive isolation between tenants.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide cryptographically isolated cognitive regions for different agent tenants.
    * Prevent "Context Smearing" and unauthorized state access between mission roots.
    * Implement tenant-specific policy enforcement and resource quotas.
    * Support seamless migration of mission state within a tenant's authorized boundaries.
* **Non-Goals:**
    * Physical hardware isolation (MTCS operates at the logical/cryptographic layer within the gateway).
    * Replacing existing per-agent session isolation (MTCS is a higher-level organizational boundary).

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise AI Infrastructure Architect
* **Primary Goal:** Host the "HR Swarm" and the "Finance Swarm" on the same MCP Any instance without risking data leakage between them.
* **The Happy Path (Tasks):**
    1. Architect defines two Tenants: `hr-tenant` and `finance-tenant`.
    2. MCP Any issues hardware-attested Tenant Identity Tokens (TIT) to each swarm.
    3. The HR agent initiates a mission; MTCS creates an isolated cognitive shard bound to `hr-tenant`.
    4. Specialist agents within the HR swarm attempt to access a tool tagged for `finance-tenant`.
    5. The MTCS isolation layer detects the tenant mismatch and interdicts the call.
    6. HR swarm state remains strictly bounded to its cognitive shard, preventing leakage into the Finance window.

## 4. Design & Architecture
* **System Flow:**
  `[Agent Request + TIT] -> [MTCS Hypervisor] -> [Tenant-Bound Shard] -> [ContextEngine] -> [Tenant-Isolated State]`
* **APIs / Interfaces:**
    * `mcpany.tenant.v1.TenantManager`
    * `rpc CreateShard(TenantIdentity) returns (ShardHandle)`
* **Data Storage/State:**
    * Tenant-specific Blackboard shards with mandatory cryptographic tagging of all KV pairs.

## 5. Alternatives Considered
* **Shared Gateway with Policy Gating (Rejected):** Insufficient protection against high-frequency "Context Smearing" and prompt injection that targets the gateway's shared memory.
* **Separate Physical Instances (Rejected):** Prohibitive resource overhead for large organizations with hundreds of small swarms.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** MTCS implements the principle of least privilege at the tenant level, ensuring absolute isolation of reasoning paths.
* **Observability:** Tenant-level telemetry and resource attribution are surfaced in the Enterprise Governance Center.

## 7. Evolutionary Changelog
* **2026-04-07:** Initial Document Creation.
