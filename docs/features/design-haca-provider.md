# Design Doc: Hardware-Attested Cost Attribution (HACA) Provider
**Status:** Draft
**Created:** 2026-07-12

## 1. Context and Scope
As AI agent swarms move from experimental playgrounds to production enterprise environments, the lack of transparent and enforceable cost attribution has become a critical barrier. In deep swarms, a single mission-root intent can spawn hundreds of subagent tool calls and reasoning branches, making it impossible to determine which specific sub-task consumed the majority of the token budget or GPU compute time.

The Hardware-Attested Cost Attribution (HACA) Provider solves this by cryptographically binding resource consumption data to the hardware-attested reasoning lineage. By utilizing TPM-signed "Economic Metadata" in every transport-layer header, HACA provides the mission-root with an immutable audit trail of every micro-cent spent, neutralizing "Economic Squatting" by rogue subagents.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a middleware layer that injects "Economic Lineage" tokens into all outbound tool-call and LLM requests.
    * Provide sub-millisecond cryptographic signing of cost-attribution metadata using local Secure Enclaves (TPM/SEP).
    * Enable real-time, mission-bound budget reconciliation for heterogeneous framework handoffs.
    * Integrate with the Reasoning-Budget Firewall (RBF) to trigger automatic capability revocation upon budget exhaustion.
* **Non-Goals:**
    * Implementing a billing or payment gateway (focus is strictly on internal swarm attribution and enforcement).
    * Managing cloud-provider quotas directly (HACA acts as the local source of truth for the mission root).

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise FinOps Lead for AI Swarms
* **Primary Goal:** Audit a failed $5,000 mission to determine which specialist agent caused the recursive token spike.
* **The Happy Path (Tasks):**
    1. The FinOps lead opens the Mission-Root Budget Registry.
    2. They select the "Project-X Refactor" mission which hit its $500 hard-limit.
    3. HACA surfaces the "Hardware-Attested Cost Tree," showing the complete lineage of spend.
    4. The lead identifies that the "Code Review Specialist" spawned 50 parallel "Refinement" branches, accounting for 80% of the spend.
    5. The lead verifies the cryptographic signatures for each branch, proving the spend was authorized by the parent intent but exceeded the local sub-quota.
    6. The lead updates the MBRQ policy to restrict "Refinement" depth for future missions.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        Agent[Agent Tool Call] -->|ARE Header| HACA[HACA Provider]
        HACA -->|Sign Attribution| TPM[Secure Enclave]
        TPM -->|Signed Token| RBF[Reasoning-Budget Firewall]
        RBF -->|Check Quota| Registry[Budget Registry]
        Registry -->|Authorized| Proxy[Egress Proxy]
    ```
* **APIs / Interfaces:**
    * `GET /haca/attribution/:mission_id`: Retrieve the hardware-attested cost tree for a mission.
    * `POST /haca/reconcile`: Aggregate sub-process lineages into a mission-wide budget report.
    * `x-mcpany-haca-token`: Standard header for propagating economic attestation across the mesh.
* **Data Storage/State:**
    * Attribution data is stored in the "Economic Shards" of the Shared KV Store, versioned by mission-root intent.

## 5. Alternatives Considered
* **Log-Based Aggregation:** Rejected due to the ease of log-spoofing by compromised subagents and the lack of cryptographic non-repudiation.
* **Cloud-Provider-Side Attribution:** Rejected because it cannot track inter-agent coordination costs or local tool-execution overhead.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Attribution tokens must be hardware-bound to the specific reasoning session. A subagent cannot "smuggle" parent-agent budget without a valid lineage signature.
* **Observability:** Cost attribution is surfaced in real-time on the Visual Attention Dashboard to provide users with "Financial Attention" maps.

## 7. Evolutionary Changelog
* **2026-07-12:** Initial Document Creation.
