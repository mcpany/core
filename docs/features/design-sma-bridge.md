# Design Doc: Swarm-Mesh Attestation (SMA) Bridge
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
With the rise of "Horizontal Swarms" (e.g., Claude Code Agent Teams) and heterogeneous meshes (OpenClaw + AutoGen), agents now frequently transition between disparate execution environments—from local developer laptops to multi-cloud production clusters. Existing session-bound tokens are insufficient for these cross-boundary shifts, leading to "Identity Squatting" where tokens from low-trust environments are replayed in high-trust meshes.

The Swarm-Mesh Attestation (SMA) Bridge aims to solve this by providing a hardware-attested, mesh-resident identity that persists across environmental boundaries.

## 2. Goals & Non-Goals
* **Goals:**
    * Issue hardware-attested (TPM/SEP) identity tokens that are cryptographically bound to a specific mission root.
    * Enable seamless identity persistence as agents move between local and cloud-based meshes.
    * Neutralize cross-mesh replay attacks by mandating origin-locked, hardware-bound handshakes.
* **Non-Goals:**
    * Replacing existing LLM provider authentication (e.g., Anthropic/OpenAI keys).
    * Managing the underlying network transport (handled by the AMT Broker).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local-First Agent Team Orchestrator
* **Primary Goal:** Securely transition a "Specialist Subagent" from a local debugging session to a high-scale production mesh without re-authenticating the mission root.
* **The Happy Path (Tasks):**
    1. The parent agent initiates an SMA-bound sub-mission on the local gateway.
    2. The local SMA Bridge issues a hardware-attested "Mesh Identity Fragment" bound to the mission root.
    3. The subagent migrates to a cloud-based node and presents the Identity Fragment to the remote SMA Bridge.
    4. The remote bridge verifies the hardware signature and mission-root lineage via a secure P2P tunnel.
    5. The subagent resumes execution with its full, attested authority.

## 4. Design & Architecture
* **System Flow:**
    * [Agent] -> [Local SMA Bridge] -> (TPM Signature) -> [Identity Fragment]
    * [Identity Fragment] -> [Remote SMA Bridge] -> [Mission Root Verification Hub]
* **APIs / Interfaces:**
    * `POST /v1/sma/mint`: Issue a new mesh-resident identity fragment.
    * `POST /v1/sma/verify`: Validate an identity fragment against its hardware-attested lineage.
* **Data Storage/State:**
    * Identity fragments are stored in a distributed, hardware-locked "Sovereignty Cache" managed by the MRCP.

## 5. Alternatives Considered
* **Short-lived JWTs:** Rejected because they are susceptible to relay attacks and do not provide hardware-bound environment attestation.
* **Centralized Auth Server:** Rejected to avoid a single point of failure and to maintain the decentralized nature of sovereign agent meshes.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All SMA fragments must be signed by a Trusted Platform Module (TPM). Re-attestation is required if the environment entropy (monitored by AEM) shifts significantly.
* **Observability:** SMA transition events are logged to the Mesh-Resident Lineage Tracker for full auditability.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
