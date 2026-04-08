# Design Doc: Recursive Lease Chaining (RLC) Provider
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
Modern agent swarms often involve deep delegation hierarchies where a specialist subagent needs to spawn its own specialist sub-agents. Current security models require either recurring HITL approvals (causing friction) or persistent over-privileged tokens (violating Zero Trust).

The Recursive Lease Chaining (RLC) Provider enables subagents to issue hardware-attested sub-leases that are cryptographically bound to the parent's mission-root lease. This ensures that authority is strictly bounded and automatically revoked across the entire chain upon mission completion.

## 2. Goals & Non-Goals
* **Goals:**
    * Facilitate deep, autonomous delegation without recurring user intervention.
    * Maintain a non-repudiable cryptographic chain of authority back to the hardware-attested mission root.
    * Enforce monotonic privilege restriction (sub-leases cannot exceed parent scope).
    * Automate recursive revocation of the entire lease tree.
* **Non-Goals:**
    * Managing cross-framework identity translation (handled by CFAT).
    * Providing long-term persistent storage of credentials (leases are ephemeral).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Securely delegate a multi-step research task to a chain of specialist agents without being prompted for every sub-task.
* **The Happy Path (Tasks):**
    1. User authorizes the primary Mission Root with a TPM-signed lease.
    2. Primary agent spawns a "Researcher" specialist and issues an RLC-chained sub-lease.
    3. "Researcher" spawns a "Data Extractor" and issues a further restricted sub-lease.
    4. Each sub-agent proves authority via its chain back to the mission root.
    5. Mission completes; the RLC Provider invalidates the entire chain across all nodes.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Mission Root Lease] --> B[RLC Chain Manager]
        B --> C[Sub-Lease Mint]
        C --> D[Sub-Agent Verification]
        D --> E[Capability Enforcement]
        E --> F[Recursive Revocation Signal]
    ```
* **APIs / Interfaces:**
    * `POST /v1/auth/lease/chain`: Issue a cryptographically bound sub-lease.
    * `GET /v1/auth/lease/verify`: Validate a lease chain against the mission root.
* **Data Storage/State:**
    * Lease chains are stored in kernel-bound memory enclaves.

## 5. Alternatives Considered
* **Token Impersonation:** Rejected as it violates identity sovereignty and provides no lineage tracking.
* **Global Session Tokens:** Rejected due to excessive blast radius if a single sub-agent is compromised.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Monotonic scope enforcement ensures that "Identity Shadowing" cannot be used to expand privileges.
* **Observability:** Visualise the "Delegation Tree" in the UI to monitor active leases.

## 7. Evolutionary Changelog
* **[2026-07-25]:** Initial Document Creation.
