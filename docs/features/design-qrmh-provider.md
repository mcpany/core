# Design Doc: Post-Quantum Mesh Handshake (QRMH) Provider
**Status:** Draft
**Created:** 2026-11-02

## 1. Context and Scope
With the rise of decentralized agent meshes (OpenClaw v4.0) and persistent agent teams, inter-agent communication often traverses diverse network environments. Current cryptographic standards, while secure today, are vulnerable to "Harvest Now, Decrypt Later" attacks by future quantum computers. To ensure the long-term auditability and sovereignty of reasoning traces, MCP Any must provide quantum-resistant security for all inter-agent coordination.

The QRMH Provider will act as the authoritative handshake broker for the Universal Agent Bus. It implements NIST-standard Post-Quantum Cryptography (PQC) to establish secure tunnels between agents, ensuring that even if the transport layer is intercepted, the semantic intent and reasoning lineage remain cryptographically shielded for the long term.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement NIST FIPS 203 (Kyber) for quantum-resistant key encapsulation.
    * Implement NIST FIPS 204 (Dilithium) for quantum-resistant digital signatures.
    * Support "Lightweight Mesh Handshakes" to minimize the latency impact of PQC.
    * Integrate with the existing hardware-attestation (TPM) workflow.
* **Non-Goals:**
    * Replacing standard TLS for non-agent traffic (e.g., UI to Server).
    * Providing a general-purpose PQC library for external applications.

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise Swarm Architect
* **Primary Goal:** Establish a quantum-secure, long-term auditable communication mesh between 50+ distributed specialist agents.
* **The Happy Path (Tasks):**
    1. Agent A initiates a coordination request to Agent B via the MCP Any Mesh Gateway.
    2. The QRMH Provider intercepts the request and performs a hardware-attested identity check.
    3. QRMH facilitates a hybrid handshake (PQC + Classical) to establish a shared secret.
    4. A quantum-resistant tunnel is established; inter-agent fragments are signed with Dilithium-based tokens.
    5. Reasoning traces are archived with a quantum-secure provenance proof.

## 4. Design & Architecture
* **System Flow:**
    [Agent A] -> (Handshake Request + TPM Proof) -> [QRMH Provider]
    [QRMH Provider] -> (KEM Ciphertext + Hybrid Secret) -> [Agent B]
    [Agent B] -> (Encrypted Confirmation) -> [Agent A]
    [Tunnel Established: AES-256-GCM + Kyber-Shared-Secret]
* **APIs / Interfaces:**
    * `POST /mesh/handshake`: Initiates a quantum-resistant handshake.
    * `GET /mesh/pqc-status`: Reports the current encryption strength and algorithm used for a tunnel.
* **Data Storage/State:**
    * Ephemeral session keys are stored in kernel-bound memory (HLES-compliant).
    * Long-term identity keys are bound to the hardware TPM.

## 5. Alternatives Considered
* **Pure Classical (X25519):** Rejected due to lack of quantum resistance.
* **Standard TLS 1.3 with PQC Extensions:** Rejected because inter-agent coordination requires more granular, mission-bound attestation than standard X.509 certs provide.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Handshakes are only initiated after a successful hardware-bound mission handshake (HADH). Failure to provide a valid mission token results in immediate quarantine.
* **Observability:** Monitor handshake latency and PQC-algorithm performance metrics via the "Fast-Path Attestation Visualizer."

## 7. Evolutionary Changelog
* **2026-11-02:** Initial Document Creation.
