# Design Doc: Headless Mission-Root (HMR) Controller
**Status:** Draft
**Created:** 2026-04-04

## 1. Context and Scope
With the rise of "Remote Control" for Claude Code and SNT for OpenClaw, AI agents are increasingly operating as persistent, headless processes in CI/CD pipelines and remote servers. The traditional "Session-Bound" security model—where trust is tied to an active terminal session—fails when the agent must persist across restarts or migrate between nodes. MCP Any needs to provide a mechanism to maintain the "Mission-Root" sovereignty in these non-interactive environments without sacrificing Zero-Trust principles.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide hardware-locked session resumption for headless agents.
    * Maintain "Chain-of-Command" lineage across process restarts.
    * Support secure "Handover" of mission-root authority between distributed nodes.
    * Integrate with TPM/Secure Enclave for non-interactive attestation.
* **Non-Goals:**
    * Provide a general-purpose long-term memory store (handled by UEG).
    * Replace existing CI/CD secret management (e.g., GitHub Secrets).

## 3. Critical User Journey (CUJ)
* **User Persona:** DevSecOps Swarm Orchestrator
* **Primary Goal:** Deploy an autonomous "Fix-and-Verify" swarm in a CI/CD pipeline that persists across multi-stage builds.
* **The Happy Path (Tasks):**
    1. Orchestrator initializes MCP Any with a TPM-signed "Mission Manifest".
    2. Headless Agent spawns and requests an HMR Lease.
    3. MCP Any issues a hardware-bound, time-limited session token.
    4. Agent performs tasks, persists state via HMR-aware Blackboard.
    5. Build stage ends; agent process terminates.
    6. Next build stage starts; new agent process resumes session using the HMR Lease and hardware attestation.

## 4. Design & Architecture
* **System Flow:**
    * `Agent Request` -> `HMR Middleware` -> `TPM/KMS Validation` -> `Session Vault`.
    * `Session Vault` stores the cryptographically signed "Mission Root" and current "Reasoning Anchor".
* **APIs / Interfaces:**
    * `POST /hmr/session/init`: Initialize a headless mission.
    * `POST /hmr/session/resume`: Resume a session with hardware attestation.
    * `GET /hmr/session/status`: Verify mission-root integrity.
* **Data Storage/State:**
    * Encrypted SQLite `blackboard.db` with HMR-specific headers for state resumption.

## 5. Alternatives Considered
* **Persistent API Keys:** Rejected because they don't provide mission-bound scoping or lineage tracking.
* **Stateless Token Exchange:** Rejected because it doesn't handle process restarts or node migration effectively in a zero-trust manner.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All resumes require a hardware-attested "Environment Proof" (e.g., Docker container hash + TPM quote).
* **Observability:** Detailed "Resumption Logs" mapping process IDs to mission-root lineage.

## 7. Evolutionary Changelog
* **2026-04-04:** Initial Document Creation.
