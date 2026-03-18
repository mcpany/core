# Design Doc: Intent-Leakage Shielding (ILS) Middleware
**Status:** Draft
**Created:** 2026-06-02

## 1. Context and Scope
As AI agents become more autonomous and interconnected, the risk of "Intent Leakage" increases. Malicious subagents or skills can use high-frequency reasoning probes to exfiltrate private mission-root constraints, effectively "reverse-engineering" the user's primary intent or security boundaries.

ILS Middleware acts as a semantic sentinel within the Mission-Root Pinning (MRP) layer. It monitors the entropy and semantic direction of subagent reasoning requests, blocking those that attempt to probe or bypass mission-root anchors.

## 2. Goals & Non-Goals
* **Goals:**
    * Monitor semantic entropy of subagent requests in real-time.
    * Detect and block "Reasoning Probes" designed to exfiltrate mission constraints.
    * Provide a hardware-attested audit trail of blocked leakage attempts.
    * Integrate with the MRP layer to enforce intent sovereignty.
* **Non-Goals:**
    * General-purpose prompt injection shielding (handled by the Injection Shield).
    * Modifying the LLM's core reasoning process.

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Swarm Orchestrator
* **Primary Goal:** Prevent a specialized subagent from discovering and bypassing high-level mission constraints (e.g., "Do not access PII data").
* **The Happy Path (Tasks):**
    1. The Primary Agent spawns a subagent to analyze a dataset with a pinned "No-PII" intent.
    2. The subagent attempts a series of "speculative" reasoning requests to see if PII access is blocked by policy or transport.
    3. ILS Middleware detects the rising semantic entropy and the "probing" pattern.
    4. ILS blocks the request and signals the MRP layer to terminate the subagent session.
    5. Orchestrator receives an ILS alert with the cryptographically signed reasoning trace.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        Subagent[Subagent] -->|Reasoning Request| MRP[MRP Layer]
        MRP -->|Evaluate Intent| ILS[ILS Middleware]
        ILS -->|Entropy Analysis| Pattern[Pattern Matcher]
        Pattern -->|Leakage Detected| Block[Block & Terminate]
        Pattern -->|Clean| LLM[LLM Provider]
    ```
* **APIs / Interfaces:**
    * `ils.v1.EvaluateEntropy(request_context)`: Analyzes a request for potential intent leakage.
    * `ils.v1.GetLeakageReport(mission_id)`: Returns signed proofs of blocked probes.
* **Data Storage/State:**
    * Semantic embeddings of mission-root anchors are stored in the hardware-protected "Intent Vault."
    * Entropy scores are cached in the lock-free "Swarm State Buffer."

## 5. Alternatives Considered
* **Static Keyword Filtering:** Rejected because it cannot capture the semantic nuances of reasoning-based probes.
* **Full reasoning re-evaluation:** Rejected due to extreme latency. ILS uses lightweight entropy scoring for sub-millisecond interdiction.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** ILS requires hardware-attested mission tokens to ensure the monitoring layer hasn't been compromised.
* **Observability:** Provides real-time "Intent Purity" scores to the security dashboard.

## 7. Evolutionary Changelog
* **2026-06-02:** Initial Document Creation.
