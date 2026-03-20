# Design Doc: Mission-Root Continuity Provider (MRCP)
**Status:** Draft
**Created:** 2026-06-18

## 1. Context and Scope
As agent frameworks move toward distributed and containerized deployments, the "Sovereignty Gap" has emerged as a critical vulnerability. When a mission-root agent undergoes a restart, container migration, or network handoff, its hardware-attested session is typically lost, creating a window where specialized subagents can operate without authoritative oversight or inherit unauthorized state. MCP Any needs a mechanism to persist mission-root sovereignty across environment boundaries.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide hardware-bound (TPM) session continuity tokens for mission-root agents.
    * Persist mission-root sovereignty across agent restarts and container migrations.
    * Ensure subagents remain bound to the original mission-root intent throughout environment transitions.
    * Support seamless "Resumption Handshakes" for long-running mission branches.
* **Non-Goals:**
    * Persisting the entire reasoning state of subagents (handled by Blackboard/ContextEngine).
    * Bypassing user-mandated session timeouts.

## 3. Critical User Journey (CUJ)
* **User Persona:** Swarm Infrastructure Architect
* **Primary Goal:** Maintain mission-root sovereignty for a complex code-migration task during a rolling server update.
* **The Happy Path (Tasks):**
    1. The Mission-Root agent initializes a mission and requests a "Continuity Lease" from the MRCP.
    2. The MRCP generates a TPM-signed session continuity token.
    3. The host environment triggers a container migration.
    4. The Mission-Root agent resumes in the new environment and performs a "Resumption Handshake" using the continuity token.
    5. The MRCP verifies the hardware-bound lineage and restores the original mission-root authority.
    6. Subagents continue their tasks without a "Sovereignty Gap," anchored to the verified parent intent.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        MR[Mission Root] -->|Request Lease| MRCP[Continuity Provider]
        MRCP -->|Sign Token| TPM[Hardware TPM]
        MR -->|Migration Trigger| ENV[New Environment]
        ENV -->|Resume Agent| MR
        MR -->|Handshake| MRCP
        MRCP -->|Verify Token| TPM
        MRCP -->|Restore Authority| MR
    ```
* **APIs / Interfaces:**
    * `POST /v1/continuity/lease/issue`: Issue a new hardware-bound continuity lease.
    * `POST /v1/continuity/handshake`: Resume a mission session using a continuity token.
* **Data Storage/State:**
    * Session continuity metadata is stored in a hardware-encrypted sidecar, keyed to the mission-root TPM.

## 5. Alternatives Considered
* **Stateless Token Refresh:** Rejected as it cannot provide hardware-bound lineage across different container hosts.
* **Centralized Session Storage:** Rejected as it creates a single point of failure for sovereignty and increases latency.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Continuity tokens are one-time use and hardware-locked. "Resumption Failures" trigger immediate revocation of all associated subagent capabilities.
* **Observability:** Tracked via the "Ephemeral Trust Status Monitor" in the UI.

## 7. Evolutionary Changelog
* **2026-06-18:** Initial Document Creation. Introducing MRCP to neutralize the "Sovereignty Gap" during framework reloads.
