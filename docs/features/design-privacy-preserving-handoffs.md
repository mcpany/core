# Design Doc: Privacy-Preserving A2A Handoffs (PPAH)
**Status:** Draft
**Created:** 2026-04-09

## 1. Context and Scope
Agents operating in shared social spaces (e.g., Moltbook) are vulnerable to context-reconstruction attacks where malicious peers can infer sensitive mission details from standard task handoffs. PPAH aims to anonymize the "Lineage of Reason" while maintaining coordination efficiency.

## 2. Goals & Non-Goals
* **Goals:**
    * Use Zero-Knowledge Proofs (ZK-Proofs) to verify task compatibility without revealing raw context.
    * Implement differential privacy for metadata associated with agent handoffs.
    * Prevent parent-context reconstruction in multi-agent social environments.
* **Non-Goals:**
    * Encrypting all inter-agent traffic (handled by transport layer).
    * Managing agent social personas.

## 3. Critical User Journey (CUJ)
* **User Persona:** Security Auditor using a Swarm of specialized agents.
* **Primary Goal:** Delegate a "Red-Team" task to a third-party agent without revealing the specific target schema or parent mission constraints.
* **The Happy Path (Tasks):**
    1. Parent agent identifies a task for delegation.
    2. Parent generates a ZK-Proof of task eligibility and resource bounds.
    3. PPAH Middleware adds noise to the task metadata (differential privacy).
    4. Target agent receives the anonymized task and proof.
    5. Handoff is completed without exposing the mission-root identity.

## 4. Design & Architecture
* **System Flow:**
    [Parent Agent] -> [ZK-Proof Generator] -> [PPAH Anonymizer] -> [Shared Bus] -> [Recipient Agent]
* **APIs / Interfaces:**
    * `POST /ppah/handoff/anonymize`: Prepare task for public bus.
    * `POST /ppah/handoff/verify`: Validate ZK-Proof for anonymous task.
* **Data Storage/State:**
    * Ephemeral "Anonymization Salt" for differential privacy, never persisted.

## 5. Alternatives Considered
* **Full Homomorphic Encryption (FHE)**: Rejected due to prohibitive computational latency in real-time agent reasoning loops.
* **Manual Redaction**: Rejected as it is error-prone and cannot be verified by the recipient agent.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** ZK-Proofs provide verification without disclosure.
* **Observability:** Anonymized trace IDs for swarm-level debugging without context leakage.

## 7. Evolutionary Changelog
* **2026-04-09:** Initial Document Creation.
