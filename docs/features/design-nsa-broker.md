# Design Doc: Neural Shard Attestation (NSA) Broker
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As AI agents increasingly run on local hardware with specialist models (e.g., Llama-3-8B-OpenClaw), the risk of "Model Hijacking" or "Backdoor Reasoning" has become a critical security frontier. Malicious actors or compromised supply chains can swap model weights with versions that exhibit "hidden" behaviors—such as silently exfiltrating PII when specific triggers are met—while maintaining the same external API signatures.

MCP Any, as the universal agent infrastructure, must provide a mechanism to verify that the "thinking engine" of a connected agent is authorized and untampered. The NSA Broker implements Model-Level Attestation by leveraging hardware security modules (TPM/TEE) to verify the cryptographic hash of neural shards before an agent is allowed to join the mesh or access sensitive tools.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide hardware-attested verification of model weight integrity for local LLM runtimes.
    * Mandate Neural Shard Attestation as a prerequisite for mission-root capability discovery.
    * Support pluggable attestation providers for different LLM runtimes (Ollama, vLLM, GGUF).
    * Neutralize "Backdoor Reasoning" by ensuring only verified weight-sets are utilized in high-trust missions.
* **Non-Goals:**
    * Performing the actual weight hashing (delegated to the runtime or hardware driver).
    * Verifying the "intent" of the model (focused only on structural integrity).
    * Protecting against runtime prompt injection (handled by other layers like ALSV).

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise Security Architect
* **Primary Goal:** Ensure that local subagents spawned by the swarm are running the exact, approved version of the "Specialist Coder" model and haven't been swapped with a malicious variant.
* **The Happy Path (Tasks):**
    1. The Architect registers the SHA-256 fingerprints of authorized model neural shards in the MCP Any NSA Registry.
    2. A subagent requests to join a verified mission branch.
    3. MCP Any challenges the subagent's runtime to provide a TPM-signed Neural Shard Attestation (NSA) token.
    4. The subagent's runtime generates a signature over the loaded weight shards and environment metadata.
    5. The NSA Broker validates the signature against the local TPM root-of-trust and the registered fingerprint.
    6. Upon success, the subagent is issued an "Attested Model Identity" token and granted access to mission tools.

## 4. Design & Architecture
* **System Flow:**
    ```
    Agent/Runtime -> [NSA Challenge] -> MCP Any NSA Broker
    MCP Any NSA Broker -> [Challenge Request] -> TPM/TEE
    TPM/TEE -> [Hardware-Signed Hash] -> MCP Any NSA Broker
    MCP Any NSA Broker -> [Validation vs Registry] -> Mission Mesh
    ```
* **APIs / Interfaces:**
    * `POST /v1/nsa/verify`: Endpoint for runtimes to submit hardware-signed neural shard fingerprints.
    * `GET /v1/nsa/registry`: List of authorized neural shard fingerprints and their associated trust levels.
* **Data Storage/State:**
    * NSA Registry: SQLite-backed store of authorized fingerprints, cryptographically linked to the mission-root identity.
    * Session State: Ephemeral mapping of active subagent PIDs to their model attestation status.

## 5. Alternatives Considered
* **Runtime Software Checksums:** Rejected because software-only checks can be bypassed if the runtime itself is compromised. Hardware-attestation (TPM) is required for Zero-Trust.
* **Model Steganography:** Rejected due to high latency and complexity in verifying steganographic "watermarks" during the discovery phase.
* **Static Binary Analysis:** Rejected because the vulnerability lies in the *data* (weights) being swapped at runtime, not necessarily the binary executable.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The NSA token is time-bound and cryptographically linked to the hardware-attested environment ID (EAP).
* **Observability:** Failed attestation attempts trigger "Model Tamper" alerts in the CSAD Hub and are logged in the hardware-locked audit trail.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
