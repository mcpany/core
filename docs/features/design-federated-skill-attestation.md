# Design Doc: Federated Skill Attestation (FSA) Hub
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
The recent proliferation of malicious skills on ClawHub (CVE-2026-25253, etc.) and the rise of "Delayed Payload" attacks have exposed a critical trust gap in the AI agent ecosystem. Current skill registries rely on static metadata or implicit trust, which are easily bypassed by sophisticated attackers.

MCP Any needs a "Federated Skill Attestation (FSA) Hub" to act as an authoritative trust oracle. This system will provide verifiable proof of a skill's behavioral and static integrity before it is allowed to execute within a mission-critical agent swarm. By federating attestations, we can leverage the collective security intelligence of multiple auditors.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a Merkle-tree based attestation store for agent skills.
    * Provide a standardized "Skill Passport" format for hardware-attested behavioral manifests.
    * Enable multi-signature validation from independent security auditors.
    * Neutralize "Delayed Payload" and "Rug Pull" supply chain attacks.
* **Non-Goals:**
    * Creating a new package manager for skills (FSA is an attestation layer, not a distribution layer).
    * Enforcing specific sandboxing technologies (FSA verifies the *claim* of safety, execution is handled by the isolation middleware).

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise Swarm Administrator
* **Primary Goal:** Securely import a 3rd party "Database Migration" skill from an untrusted marketplace.
* **The Happy Path (Tasks):**
    1. Administrator provides the skill's manifest URL to the MCP Any FSA Hub.
    2. FSA Hub queries federated registries for the skill's hash.
    3. FSA Hub retrieves a "Skill Passport" containing hardware-attested behavioral profiles (e.g., "Only accesses /tmp", "No outbound network").
    4. FSA Hub verifies the multi-signature audit from trusted providers (e.g., STSS, VIRUSTOTAL-AI).
    5. FSA Hub issues a local "Capability Lease" cryptographically bound to the mission root.
    6. The agent executes the skill within the verified behavioral boundaries.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Agent Registry] -->|Manifest| B(FSA Hub)
        B -->|Hash Query| C{Federated Registries}
        C -->|Skill Passport| B
        B -->|Verify Signatures| D[Policy Engine]
        D -->|Issue Lease| E[Execution Environment]
    ```
* **APIs / Interfaces:**
    * `GET /v1/attestations/{skill_hash}`: Retrieve verified passports and auditor signatures.
    * `POST /v1/passport/verify`: Submit a local skill for federated attestation verification.
* **Data Storage/State:**
    * Persistent SQLite store for local attestation caches and "Skill Passport" metadata.
    * Merkle Tree root pinned to hardware (TPM) to ensure cache integrity.

## 5. Alternatives Considered
* **Centralized Registry**: Rejected because it creates a single point of failure and doesn't scale across disparate agent frameworks.
* **Static Analysis Only**: Rejected because sophisticated "Delayed Payloads" only trigger under specific environmental conditions, necessitating behavioral auditing.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All passports must be hardware-signed (TPM). Revocation lists (ARL) are checked in real-time.
* **Observability:** Audit logs capture every attestation check, including confidence scores from different auditors.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
