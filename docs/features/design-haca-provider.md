# Design Doc: Hardware-Attested Cost Attribution (HACA) Provider
**Status:** Draft
**Created:** 2026-07-12

## 1. Context and Scope
As agent swarms become more autonomous and deep, "Economic Squatting" has emerged as a significant threat. Malicious or poorly optimized subagents can exhaust a parent mission's token and compute budget through opaque recursive calls or shadow delegations. Existing cost tracking is often coarse-grained and easily spoofed by sub-processes.

The Hardware-Attested Cost Attribution (HACA) Provider introduces a cryptographically secure layer for economic accountability. By leveraging TPM-bound monotonic counters and reasoning-lineage tokens, HACA ensures that every micro-cent of token spend and every millisecond of GPU time is attributed to a specific, verified branch of the mission root. This enables precise budget enforcement and prevents resource exhaustion in high-density meshes.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a hardware-bound attribution token that follows every tool call and reasoning fragment.
    * Provide real-time, tamper-proof cost reporting for deep subagent lineages.
    * Integrate with the Reasoning-Budget Firewall (RBF) for sub-millisecond budget interdiction.
    * Support Gemini CLI's ARE v1.9 standard for cost-metadata propagation.
* **Non-Goals:**
    * Implementing a billing or payment system (HACA provides the *attribution*, not the *transaction*).
    * General-purpose system profiling (focus is strictly on agentic resource consumption).

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise AI Ops Manager
* **Primary Goal:** Identify which specific sub-specialist in a 50-agent swarm is responsible for a sudden $2,000 token spike.
* **The Happy Path (Tasks):**
    1. A large-scale software auditing mission is running in production.
    2. The Reasoning-Budget Firewall detects a cost-burn rate exceeding the mission-root threshold.
    3. The manager opens the HACA Dashboard.
    4. They drill down into the "Refactoring" intent branch and see a hierarchical tree of spend.
    5. HACA identifies that the "Legacy Code Archaeologist" subagent is stuck in a $0.50/turn reasoning loop due to an unoptimized prompt.
    6. The manager revokes the subagent's budget lease via the HACA interface, halting the drain immediately.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        Agent[Subagent] -->|Tool Call + HACA Token| Gateway[MCP Any Gateway]
        Gateway -->|Verify Lineage| HACA[HACA Provider]
        HACA -->|Query TPM| TPM[Hardware TPM]
        HACA -->|Check Budget| RBF[Reasoning-Budget Firewall]
        RBF -->|Authorized| Upstream[Model Provider]
        Upstream -->|Usage Stats| HACA
        HACA -->|Update Lineage Cost| Registry[Cost Registry]
    ```
* **APIs / Interfaces:**
    * `GET /haca/attribution/{mission_root_id}`: Hierarchical spend report for a mission.
    * `POST /haca/lease/reclaim`: Forcefully reclaim unused token leases from a sub-lineage.
    * `header: x-mcp-haca-token`: TPM-signed attribution token.
* **Data Storage/State:**
    * Hierarchical cost-state is stored in a mission-bound SQLite sidecar, with periodic hardware-attested snapshots for non-repudiation.

## 5. Alternatives Considered
* **Log-Based Attribution:** Rejected because logs are post-facto and can be spoofed or deleted by a compromised agent. HACA provides *pre-execution* enforcement.
* **Standard JWT Scoping:** Rejected due to lack of hardware binding, making tokens reusable across disparate physical environments.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** HACA tokens must be non-reusable and cryptographically bound to the hardware session.
* **Observability:** Integrates with the "Economic Attribution Viewer" in the UI for real-time visualization.

## 7. Evolutionary Changelog
* **2026-07-12:** Initial Document Creation. Integration with UACO v3.6 Recursive Resource Reclamation.
