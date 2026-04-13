# Design Doc: Cognitive Boundary Enforcer (CBE)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
The emergence of "Cross-Reasoning Contamination" (CRC) in Gemini CLI and OpenClaw confirms that logical isolation at the transport layer is insufficient. Malicious subagents can influence parent decision-making logic via side-channels in shared caches or process environments.

The Cognitive Boundary Enforcer (CBE) is designed to provide physical, kernel-level isolation for subagent reasoning paths. By utilizing gVisor-based sandboxing and Distributed Memory Enclaves (DME), CBE ensures that a specialist's "cognitive workspace" is strictly segmented from its peers and parent.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement gVisor-based sandbox isolation for all subagent execution.
    * Segment shared reasoning buffers into hardware-locked Distributed Memory Enclaves.
    * Neutralize side-channel "Reasoning Leakage" between specialist agents.
    * Enforce mission-bound boundaries for all in-memory state fragments.
* **Non-Goals:**
    * Isolating network traffic (handled by Isolated Named-Pipes).
    * Managing agent-to-user UI rendering (handled by A2UI).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Developer
* **Primary Goal:** Prevent a low-trust "Web Search" specialist from accessing the memory region of a high-trust "Database Admin" agent.
* **The Happy Path (Tasks):**
    1. The Mission-Root spawns a Web Search agent and a DB Admin agent.
    2. CBE assigns each agent a unique gVisor sandbox and a non-overlapping DME segment.
    3. The Web Search agent attempts to read a "Side-Channel Buffer" used by the DB Admin.
    4. gVisor intercepts the unauthorized memory access.
    5. CBE triggers a "Sovereignty Breach" alert and terminates the Web Search session.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph LR
    ParentAgent -->|Spawn Subagent| CBE
    CBE -->|Provision| gVisorSandbox
    CBE -->|Allocate| DMESegment
    Subagent -->|Reasoning| DMESegment
    gVisorSandbox -->|Intercept Cross-Access| MemoryWatchdog
    MemoryWatchdog -->|Alert| CBE
    ```
* **APIs / Interfaces:**
    * Internal: `cbe.ProvisionSandbox(agentID, missionID)`
    * Internal: `cbe.AllocateEnclave(agentID, size)`
    * Management: `GET /v1/cbe/integrity`: Returns the attestation status of all isolated regions.
* **Data Storage/State:**
    * Metadata about sandbox-to-enclave mappings is stored in secure, kernel-bound memory.

## 5. Alternatives Considered
* **Standard Docker Isolation:** Rejected due to shared kernel surface area, which is vulnerable to "Reasoning-Path Escapes."
* **Software-Only Redaction:** Rejected because it doesn't prevent lower-level side-channel probes (e.g., timing).

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Relies on hardware-attested (TPM) proofs for DME integrity.
* **Observability:** Provides real-time "Isolation Health" scores for the Service Mesh Topology Monitor.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
