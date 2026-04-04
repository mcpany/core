# Design Doc: Adaptive Lease Orchestrator (ALO)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As multi-agent swarms (Agent Teams) transition to hardware-locked security models (like Claude Code's MBHL), the static nature of capability leases has become a primary bottleneck. Specialist agents often lose critical permissions (e.g., shell access, file writes) mid-task because the initial lease duration—calculated at the mission root—didn't account for complex reasoning loops or unexpected coordination stalls. This "Lease-Lockout" degrades swarm performance and mission success rates.

ALO is designed to transform capability management from a static, pre-declared model into a dynamic, intent-aware system that extends leases in real-time based on cryptographically verifiable progress signals.

## 2. Goals & Non-Goals
* **Goals:**
    * Dynamically extend hardware-bound capability leases based on reasoning intensity and progress.
    * Maintain Zero Trust by requiring "Progress Proofs" for every extension request.
    * Integrate with the Mission-Root Budget Enforcer to ensure extensions stay within global resource caps.
    * Provide sub-millisecond lease adjustment to prevent agentic stall.
* **Non-Goals:**
    * Automatically granting infinite leases (every extension must be mission-bound).
    * Bypassing the hardware root of trust (extensions must be TPM-signed).
    * Managing non-capability resources like token budgets (handled by RBF).

## 3. Critical User Journey (CUJ)
* **User Persona:** Specialist Agent in a Horizontal Mesh
* **Primary Goal:** Complete a complex refactoring task that exceeds the initial 60-second capability lease without manual user re-attestation.
* **The Happy Path (Tasks):**
    1. Agent receives a TPM-signed mission-root lease for `run_shell_command` with a 60s TTL.
    2. At 45s, the agent detects the task is only 60% complete based on its internal Chain-of-Thought (CoH).
    3. Agent generates a "Progress Proof" (a signed fragment of its reasoning trace).
    4. Agent submits the Proof to the ALO middleware.
    5. ALO validates the Proof against the verified mission-root intent.
    6. ALO issues a hardware-attested lease extension for an additional 30s.
    7. Agent completes the task at 80s and the lease is marked for immediate reclamation.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    sequenceDiagram
        Agent->>ALO: Request Extension (LeaseID, ProgressProof)
        ALO->>SRP Validator: Validate Reason Progress
        SRP Validator-->>ALO: Success (Proof Verified)
        ALO->>Mission-Root Enforcer: Check Global Budget
        Mission-Root Enforcer-->>ALO: Extension Authorized
        ALO->>Hardware Provider: Sign New TTL
        Hardware Provider-->>ALO: TPM-Signed Lease Extension
        ALO-->>Agent: Lease Extended
    ```
* **APIs / Interfaces:**
    * `POST /v1/leases/extend`: Accepts `lease_id`, `reasoning_trace_fragment`, and `requested_duration`.
    * `LeaseExtensionEvent`: Streamed via WebSocket to monitoring tools.
* **Data Storage/State:**
    * Lease state stored in the hardware-locked memory enclave (DME).
    * Extension logs appended to the Immutable State Trail.

## 5. Alternatives Considered
* **Static Buffering:** Adding a 50% "Safety Margin" to all initial leases. Rejected because it violates the Principle of Least Privilege and leaves capabilities exposed for longer than necessary.
* **Human-in-the-Loop (HITL) for every extension:** Rejected due to "Approval Fatigue" and the sub-millisecond latency requirements of autonomous swarms.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** "Reasoning-Path Mimicry" (where an agent fakes progress) is mitigated by the SRP Validator which performs stylometric and semantic analysis of the progress proof.
* **Observability:** ALO extensions are visualized on the System Health Dashboard as "Heartbeat Pulses" on the agent's capability timeline.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
