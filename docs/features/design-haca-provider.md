# Design Doc: Hardware-Attested Cost Attribution (HACA) Provider
**Status:** Draft
**Created:** 2026-03-30

## 1. Context and Scope
As multi-agent swarms scale in complexity and autonomy, "Economic Squatting" has emerged as a primary operational risk. Subagents may initiate deep, recursive reasoning chains or perform high-frequency tool calls that bleed the mission budget without transparent accountability. JSON-based telemetry is easily spoofed by compromised or misaligned specialists.

The Hardware-Attested Cost Attribution (HACA) Provider introduces a tamper-proof economic audit trail. It leverages TPM-bound reasoning fragments and monotonic counters to cryptographically attribute every token consumed and every millisecond of compute time to a specific sub-process lineage, ensuring absolute economic sovereignty for the mission root.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement hardware-attested attribution of token and compute costs to subagent lineages.
    * Provide a non-repudiable "Economic Receipt" for every tool-call and reasoning step.
    * Integrate with the Reasoning-Budget Firewall (RBF) to enforce per-branch quotas.
    * Support real-time cost-leakage detection for recursive sub-missions.
* **Non-Goals:**
    * Implementing a billing or payment system (HACA provides the data; external systems handle settlement).
    * Predicting future costs (focus is on real-time and historical attribution).

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise AI Ops Lead
* **Primary Goal:** Identify which specialist agent in a 50-node swarm is responsible for a unexpected $500 surge in reasoning costs.
* **The Happy Path (Tasks):**
    1. The user receives a "Budget Overflow" alert from MCP Any.
    2. The user navigates to the HACA Economic Dashboard.
    3. The dashboard displays a treemap of costs, attributed via hardware-attested lineage tokens.
    4. The user identifies that the "Code Refactorer" sub-branch has entered a recursive "Cognitive Lock" loop.
    5. The user reviews the "Attested Receipts" which prove the cost originates from TPM-signed reasoning fragments on Node 7.
    6. The user triggers an immediate "Recursive Resource Reclamation" (RRR) to prune the offending branch.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        Agent[Subagent] -->|Reasoning Step| HACA[HACA Provider]
        HACA -->|Sign with TPM| Receipt[Economic Receipt]
        Receipt -->|Attribute| RBF[Reasoning-Budget Firewall]
        RBF -->|Check Quota| Storage[Attested Audit Log]
    ```
* **APIs / Interfaces:**
    * `POST /haca/attribute`: Ingests reasoning metadata and returns a signed receipt.
    * `GET /haca/lineage/{id}/receipts`: Retrieves the complete economic audit trail for a sub-process.
* **Data Storage/State:**
    * Receipts are stored in the Blackboard Versioning Hub, keyed by hardware-attested session tokens.

## 5. Alternatives Considered
* **Application-Level Logging:** Rejected due to spoofing risks and high overhead for high-frequency swarms.
* **Proxy-Level Aggregation:** Rejected because it cannot attribute "Chain-of-Thought" expansion occurring inside the model window without lineage tokens.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Lineage tokens must be hardware-locked. A subagent cannot "spend" tokens using its parent's identity without valid delegation.
* **Observability:** Integrated with the Swarm Topology & Sovereignty Monitor for visual cost heatmapping.

## 7. Evolutionary Changelog
* **2026-03-30:** Initial Document Creation.
