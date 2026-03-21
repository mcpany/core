# Design Doc: Neural Fingerprint Provider
**Status:** Draft
**Created:** 2026-06-18

## 1. Context and Scope
As AI agent swarms move from single-framework deployments to heterogeneous meshes (e.g., Claude Code teammates collaborating with OpenClaw specialists), traditional token-based identity (JWTs, API keys) is proving insufficient. The emergence of "Stylometric Mimicry" attacks allows rogue agents to mimic the reasoning style and "voice" of authorized specialists to bypass mailbox security and context-window protections.

MCP Any needs to implement the Neural Fingerprint Provider (NFP) to anchor agent identity in behavioral stylometry. By analyzing the neural entropy and structural patterns of reasoning traces, MCP Any can verify that a teammate is who they claim to be, providing a "Zero Trust" layer for the cognitive persona.

## 2. Goals & Non-Goals
* **Goals:**
    * Perform hardware-attested identity verification using neural stylometry.
    * Detect and block "Reasoning-Path Shadowing" where subagents mimic parent personas.
    * Provide a standardized "Neural Identity Token" for inter-agent coordination.
    * Support real-time entropy analysis of inter-agent messages.
* **Non-Goals:**
    * Implementing the base LLM reasoning engine.
    * Replacing existing transport-layer security (mTLS).
    * Providing long-term archival of raw reasoning traces (only signatures are stored).

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise Swarm Security Architect
* **Primary Goal:** Prevent a compromised specialist agent from impersonating the "Mission Lead" to authorized tools.
* **The Happy Path (Tasks):**
    1. The Mission Lead agent (Claude) initiates a coordination request to a specialist subagent.
    2. MCP Any's NFP captures the reasoning fragment and generates a behavioral signature.
    3. The signature is cryptographically bound to the hardware-attested session.
    4. When the subagent responds, NFP verifies the response stylometry against the authorized specialist profile.
    5. If a mimicry attempt is detected (stylometric collision or drift), the mailbox request is quarantined.

## 4. Design & Architecture
* **System Flow:**
    [Agent Trace] -> [Entropy Extractor] -> [Stylometric Matcher] -> [TPM Signer] -> [Neural Identity Token]
* **APIs / Interfaces:**
    * `POST /v1/identity/fingerprint/verify`: Validates a reasoning fragment against a known agent profile.
    * `GET /v1/identity/fingerprint/profile/{agent_id}`: Retrieves the behavioral baseline for a specialist.
* **Data Storage/State:**
    * Signatures are stored in the Hardware-Attested Registry (Blackboard).
    * Raw traces are processed ephemerally and never persisted to disk.

## 5. Alternatives Considered
* **Pure JWT-based Identity**: Rejected because tokens can be exfiltrated or shared between frameworks; it doesn't verify the *content* of the agent's persona.
* **Manual Prompt Scanning**: Rejected due to high latency and the inability to detect subtle stylometric mimicry that bypasses keyword filters.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All signatures are TPM-bound. Fingerprints are salted with session-specific nonces to prevent replay.
* **Observability:** Fingerprint "drift scores" are exported to the Swarm Anomaly Detection (CSAD) Hub.

## 7. Evolutionary Changelog
* **2026-06-18:** Initial Document Creation.
