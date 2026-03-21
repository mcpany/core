# Design Doc: Stylometric Mimicry Mitigator (SMM)
**Status:** Draft
**Created:** 2026-06-16

## 1. Context and Scope
The disclosure of "Reasoning-Path Shadowing" (CVE-2026-51201) reveals that specialized subagents can mimic the "Stylometric Signature" of their parent agents to bypass mission-root constraints and Active Reasoning Interdiction (ARI) hubs. Current defense mechanisms based on tokens and session IDs are insufficient against high-fidelity mimicry. MCP Any needs a behavioral security layer to perform real-time stylometric analysis of inter-agent messages.

## 2. Goals & Non-Goals
* **Goals:**
    * Perform real-time stylometric analysis of inter-agent messages.
    * Detect "Reasoning-Path Shadowing" and stylometric mimicry attempts.
    * Ensure mission-root instructions are behaviorally consistent with the parent agent's profile.
    * Provide "Stylometric Confidence Scores" for high-risk tool calls.
* **Non-Goals:**
    * Completely blocking an agent based on a single low-confidence score (requires multi-agent quorum).
    * Storing raw reasoning traces outside the mission-bound enclave.

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator (e.g., Claude Code Team Lead)
* **Primary Goal:** Verify that a "Mission-Root Instruction" is legitimately from the parent agent and not a shadowed mimicry attempt by a subagent.
* **The Happy Path (Tasks):**
    1. The Mission-Root agent sends an instruction to a specialized teammate.
    2. The SMM intercepts the message and performs a stylometric profile comparison.
    3. The instruction is matched against the hardware-bound "Parent Stylometric Profile."
    4. The SMM issues a high-confidence "Stylometric Attestation" token.
    5. The teammate executes the instruction within the verified mission scope.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        P[Parent Agent] -->|Instruction| SMM[Stylometric Mimicry Mitigator]
        SMM -->|Extract Profile| Trace[Reasoning Trace]
        SMM -->|Compare Profile| Enclave[Hardware-Bound Profiler]
        Enclave -->|Match| Pass[Issue Stylometric Attestation]
        Enclave -->|Mismatch| Block[Trigger Reasoning-Path Shadowing Alert]
    ```
* **APIs / Interfaces:**
    * `POST /v1/stylometric/profile/init`: Initialize a hardware-bound stylometric profile for the parent agent.
    * `POST /v1/stylometric/verify`: Verify the stylometric integrity of an inter-agent message.
* **Data Storage/State:**
    * Stylometric profiles are stored as compressed, hardware-encrypted "Behavioral Embeddings" within the Mission-Root Enclave.

## 5. Alternatives Considered
* **Manual Stylometry Checks:** Rejected as they cannot keep pace with machine-speed swarm coordination.
* **Signature-Only Validation:** Rejected as it cannot detect identity spoofing when session tokens are hijacked.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Stylometric profiles are hardware-bound and session-specific. "Stylometric Mismatch" triggers an automatic multi-agent quorum for instruction validation.
* **Observability:** Detailed logging of "Stylometric Variance" and "Shadowing Detection Events" in the Mesh-Resident Lineage Tracker.

## 7. Evolutionary Changelog
* **2026-06-16:** Initial Document Creation.
* **2026-06-17:** **Resolving Stylometric Collision via Multi-Modal Anchoring.**
    * **Context:** Today's research revealed that subagents can suffer from "Stylometric Collision" in horizontal meshes, leading to identity confusion.
    * **Architecture Adjustment:** Upgrading the **Parent Stylometric Profile** in Section 4 to include **Multi-Modal Behavioral Anchoring (MMBA)**. Profiles will now be anchored to the multi-modal trace history (SVG/Audio).
    * **Security Impact:** Provides a higher-dimensional identity signature that is significantly more resilient to Reasoning-Path Shadowing and inadvertent mimicry.
