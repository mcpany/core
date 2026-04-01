# Design Doc: Intent-Chain Sovereignty (ICS) Validator
**Status:** Draft
**Created:** 2026-07-23

## 1. Context and Scope
As AI agent swarms grow in complexity, the risk of "Intent Ghosting" increases. When a mission root delegates tasks to subagents, and those subagents further delegate to specialists, the original constraints and security boundaries often become diluted or lost. Existing session-based security is insufficient for protecting the hierarchical integrity of a multi-hop mission.

The **Intent-Chain Sovereignty (ICS) Validator** is an infrastructure-level service for MCP Any that cryptographically anchors every delegated sub-intent to the user's primary mission root. It ensures that any action taken by a specialist agent can be traced back through a hardware-attested chain of lineage to the original authorized intent.

## 2. Goals & Non-Goals
* **Goals:**
    * provide a standardized protocol for hierarchical intent anchoring (HIA).
    * Validate the complete cryptographic lineage of a tool call back to the mission root.
    * Support hardware-attested (TPM/Secure Enclave) intent fragments.
    * Enable sub-millisecond lineage verification during high-frequency teammate coordination.
* **Non-Goals:**
    * modifying the internal reasoning of the agents.
    * Providing long-term storage for reasoning traces (handled by the SRM provider).
    * enforcing budgetary constraints (handled by the Reasoning-Budget Firewall).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Swarm Orchestrator
* **Primary Goal:** Ensure that a "Security Auditor" subagent cannot be coerced by a "Developer" subagent into approving a malicious PR that diverges from the mission root.
* **The Happy Path (Tasks):**
    1. The User initiates a mission with a TPM-signed "Mission Root Manifest."
    2. The Lead Agent spawns a "Security Auditor" subagent, attaching a signed "Intent Fragment" linked to the root.
    3. The Auditor attempts a tool call to `approve_pr`.
    4. The ICS Validator intercepts the call and verifies the cryptographic chain from the Auditor's fragment back to the Mission Root.
    5. The Validator confirms the `approve_pr` action aligns with the authorized scope in the root manifest.
    6. The tool call is authorized.

## 4. Design & Architecture
* **System Flow:**
    `Agent` -> `ICS Middleware` -> `Lineage Store` -> `TPM Validator` -> `Policy Engine`.
* **APIs / Interfaces:**
    * `x-mcp-intent-chain`: Header containing the hash-chained intent fragments.
    * `/v1/intent/anchor`: Endpoint for minting new sub-intent fragments.
* **Data Storage/State:**
    * Ephemeral, hardware-protected cache of verified intent lineages.
    * Integration with the Shared KV Store (Blackboard) for intent-bound state isolation.

## 5. Alternatives Considered
* **Flat Session Tokens**: Rejected as they do not capture hierarchical delegation or provide protection against lateral "Intent Hijacking."
* **Centralized Intent Registry**: Rejected due to latency and the "Supervisor Bottleneck"; decentralized cryptographic anchoring scales better with autonomous swarms.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All intent fragments are hardware-signed, preventing subagents from spoofing parent authority.
* **Observability:** The "Intent-Chain Visualization Hub" provides real-time visibility into the mesh's hierarchical sovereignty.

## 7. Evolutionary Changelog
* **2026-07-23:** Initial Document Creation.
