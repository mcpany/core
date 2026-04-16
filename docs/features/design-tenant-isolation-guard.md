# Design Doc: Tenant Isolation Guard (TIG)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As enterprise adoption of MCP swarms accelerates, a critical vulnerability known as "Context-Echoing" has emerged (highlighted by Asana's 2026 isolation flaw). Multiple tenants often share the same underlying LLM model caches and coordination shards, leading to data leakage where one tenant's private mission context becomes visible or "echoed" in another's reasoning loop.

The Tenant Isolation Guard (TIG) is required to act as the authoritative "Tenant Broker," providing hardware-attested boundary isolation for all enterprise data fragments and inter-agent coordination.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide hardware-attested isolation for tenant-specific context fragments.
    * Neutralize "Context-Echoing" by enforcing tenant-bound encryption for all shared coordination shards.
    * Implement a multi-tenant session manager that cryptographically separates mission-root lineages.
    * Integrate with ZKSA for verifiable boundary integrity without context exposure.
* **Non-Goals:**
    * Managing tenant billing or subscription status.
    * Providing network-layer VPC isolation (this layer sits above the network).
    * Modifying upstream LLM provider cache logic (TIG sanitizes inputs/outputs to prevent cross-tenant cache hits).

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise Security Architect
* **Primary Goal:** Ensure that financial data from the "Accounting" tenant is never leaked to the "Sales" tenant's agent swarm.
* **The Happy Path (Tasks):**
    1. Architect defines "Accounting" and "Sales" tenant IDs in MCP Any.
    2. Accounting agent initiates a tool call; TIG tags the request with a hardware-attested Tenant Token.
    3. TIG encrypts the resulting context shard with a key unique to the "Accounting" tenant and the hardware enclave.
    4. Sales agent attempts to discover tools or access coordination shards; TIG intercepts and verifies the lack of matching Tenant Tokens.
    5. Sales agent is restricted to its own isolated shards, with zero visibility into Accounting data.
    6. Auditor uses ZKSA to verify that no cross-tenant "echoes" occurred during the session.

## 4. Design & Architecture
* **System Flow:**
    `Agent Request` -> `Tenant Identification` -> `Hardware-Attested Token Issuance` -> `Encrypted Shard Storage` -> `Zero-Knowledge Verification`
* **APIs / Interfaces:**
    * `tig.IssueTenantToken(tenantID, hardwareID) -> TenantToken`: Mints a hardware-locked isolation token.
    * `tig.WrapShard(tenantToken, contextFragment) -> EncryptedShard`: Binds a fragment to a specific tenant boundary.
    * `tig.VerifyIsolationQuorum(tenantA, tenantB) -> ZKProof`: Generates a proof that no data leaked between tenants.
* **Data Storage/State:**
    * **Tenant Key Vault**: HSM-backed storage for per-tenant encryption keys.
    * **Isolation Registry**: Real-time mapping of active sessions to tenant boundaries.

## 5. Alternatives Considered
* **Namespace Isolation (Logical)**: Rejected as logical separation is vulnerable to side-channel timing attacks and "echoing" in model attention maps.
* **Full Air-Gapped Instances**: Too expensive and complex to manage for thousands of enterprise tenants. TIG provides "Enclave-Level" isolation with the efficiency of shared infrastructure.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: All inter-tenant interactions are blocked by default. Keys are never exposed to the agent reasoning engine.
* **Observability**: Real-time "Tenant Leakage" alerts in the UI dashboard and automated ZKSA audit logs.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
