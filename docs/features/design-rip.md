# Design Doc: Regulatory Inventory Provider (RIP)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
The enforcement of the EU AI Act and similar global regulations mandates that organizations maintain a comprehensive, auditable inventory of all AI systems, models, and third-party plugins in use. In dynamic agent swarms, where agents can autonomously discover and graft new skills (MCP servers), manual inventory is impossible.

MCP Any requires the Regulatory Inventory Provider (RIP) to act as the authoritative "Compliance Mint." It provides a hardware-attested, real-time registry of the entire agentic supply chain, satisfying regulatory transparency and risk assessment requirements.

## 2. Goals & Non-Goals
* **Goals:**
    * Maintain a real-time, hardware-attested inventory of all active models, MCP servers, and subagent frameworks.
    * Track "Model Lineage" and "Plugin Provenance" for every mission-root.
    * Generate EU AI Act compliant compliance reports (PDF/JSON) on demand.
    * Automate the "Capability-to-Risk" mapping for enterprise swarms.
* **Non-Goals:**
    * Providing general-purpose model performance monitoring (handled by Telemetry).
    * Enforcing runtime policies (handled by the Policy Firewall).

## 3. Critical User Journey (CUJ)
* **User Persona:** Compliance Officer
* **Primary Goal:** Generate an audit report for a production swarm to verify that no "High-Risk" un-sanitized third-party tools were used during a specific period.
* **The Happy Path (Tasks):**
    1. Compliance Officer requests a "Supply Chain Report" via the MCP Any RIP dashboard.
    2. RIP queries the Blackboard for all hardware-attested capability cards used in the last 30 days.
    3. RIP cross-references these cards with the Verified Skill Registry to confirm provenance.
    4. RIP generates a TPM-signed report listing model versions, MCP server hashes, and their associated risk tiers.
    5. Officer submits the signed report to the regulatory body.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Agent Discovery] --> B[RIP Interceptor]
        B --> C[Blackboard / State]
        D[Compliance Request] --> E[RIP Generator]
        E --> F[Hardware Enclave / Signing]
        F --> G[Attested Report]
    ```
* **APIs / Interfaces:**
    * `/v1/compliance/inventory`: Retrieve the current live inventory.
    * `/v1/compliance/report/generate`: Trigger a signed report generation.
* **Data Storage/State:**
    * Inventory snapshots are stored as "Compliance Anchors" in the local SQLite vault, cryptographically bound to the host TPM.

## 5. Alternatives Considered
* **Static Config Audit:** Rejected because it misses autonomously discovered tools and dynamic skill grafting.
* **Log Aggregation (SIEM):** Rejected as logs lack the hardware-attestation and structured metadata required for EU AI Act compliance.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** RIP itself is origin-locked and requires MFA for report generation.
* **Observability:** RIP sync latency is monitored in the Swarm Coherence Dashboard.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
