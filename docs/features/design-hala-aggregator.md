# Design Doc: Hardware-Attested Lease Aggregator (HALA)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
The emergence of high-density Agent Teams (e.g., Claude Code, OpenClaw swarms) has led to a critical performance bottleneck known as "Attestation Storms." In these environments, specialist agents may perform hundreds of tool calls per minute, each requiring a unique hardware-attested (TPM/SEP) capability lease. The per-call overhead of individual hardware signatures is causing significant kernel-level latency and cognitive stall.

The Hardware-Attested Lease Aggregator (HALA) is designed to solve this by bundling multiple task-bound capability leases into a single hardware-attested aggregate token, drastically reducing the number of physical attestation calls required.

## 2. Goals & Non-Goals
* **Goals:**
    * Aggregate multiple subagent capability leases into a single TPM-signed bundle.
    * Reduce per-call attestation latency by at least 40% in high-density swarms.
    * Maintain granular, mission-bound authorization within the aggregate token.
    * Provide a fallback mechanism for individual lease attestation if aggregation fails.
* **Non-Goals:**
    * Replacing the underlying TPM/SEP hardware requirements.
    * Managing the task delegation logic itself; HALA is a security transport optimization.
    * Aggregating leases across disparate mission roots (bundles must be mission-scoped).

## 3. Critical User Journey (CUJ)
* **User Persona:** High-Density Swarm Orchestrator
* **Primary Goal:** Efficiently authorize 50+ tool calls from 5 specialist agents without incurring kernel-level TPM bottlenecks.
* **The Happy Path (Tasks):**
    1. Orchestrator identifies a burst of upcoming tasks for a specific mission branch.
    2. Orchestrator requests an "Aggregate Lease" from the HALA service.
    3. HALA collects the individual task manifests and generates a single bundle.
    4. The local TPM signs the bundle hash.
    5. The resulting HALA token is distributed to the specialist agents.
    6. Agents present the HALA token to the Gateway for tool execution.
    7. Gateway verifies the single signature and the inclusion of the specific task in the bundle.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Orchestrator] -->|Task Manifests| B[HALA Service]
        B -->|Hash Bundle| C[TPM/Hardware]
        C -->|Aggregate Signature| B
        B -->|HALA Token| D[Specialist Agents]
        D -->|HALA Token + Tool Call| E[Gateway]
        E -->|Verify Signature & Membership| F[Execute Tool]
    ```
* **APIs / Interfaces:**
    * `hala.CreateBundle(manifests[]) -> HalaToken`: Generates an aggregate lease bundle.
    * `hala.VerifyLease(halaToken, taskID) -> bool`: Validates a specific lease within a bundle.
    * `hala.ExtendBundle(halaToken, newManifests[]) -> HalaToken`: (Optional) Optimistically adds new leases to an active bundle.
* **Data Storage/State:**
    * **Bundle Registry:** Short-lived in-memory cache of active aggregate tokens.
    * **Nonce Manager:** Ensures monotonic integrity of bundled leases to prevent replay attacks.

## 5. Alternatives Considered
* **Persistent Trust Leases (LFTA):** Already implemented, but LFTA still requires per-session hardware re-attestation. HALA optimizes the *burst* phase of high-frequency sessions.
* **Software-Only Aggregation:** Rejected because it violates the "Hardware-Attested" core pillar of MCP Any security.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** HALA tokens are cryptographically bound to the mission root. Revoking the mission root automatically invalidates the HALA bundle.
* **Observability:** Metrics for "Aggregation Ratio" (Leases per Signature) and "TPM Contention Rate" will be surfaced in the UI.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
