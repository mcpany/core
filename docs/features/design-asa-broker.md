# Design Doc: Autonomous Sovereignty Arbiter (ASA) Broker
**Status:** Draft
**Created:** 2026-11-02

## 1. Context and Scope
As AI agent swarms transition from hierarchical, single-framework control (e.g., just Claude Code) to horizontal, heterogeneous meshes (Claude, OpenClaw, and AutoGen teammates), a critical bottleneck has emerged: **Instruction Collision**. Specialist agents often receive conflicting tasks or priority signals from different supervisors, leading to "Cognitive Stall" and resource exhaustion.

The Autonomous Sovereignty Arbiter (ASA) Broker is designed to move MCP Any from a passive message relay to an active, constitutional mediator. It allows a mission root to define a "Mission Constitution" that subagents can use to autonomously resolve conflicts without round-tripping to a human or a high-latency supervisor.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Provide a hardware-attested registry for "Mission Constitutions."
    *   Enable specialist agents to autonomously resolve instruction conflicts based on constitutional priority.
    *   Reduce MTTC (Mean Time to Coordinate) in heterogeneous meshes by 50%+.
    *   Ensure non-repudiable arbitration logs for forensic auditing.
*   **Non-Goals:**
    *   Replacing the primary reasoning engine of the agents.
    *   Defining the specific ethics of the agents (the user defines the constitution).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Heterogeneous Swarm Orchestrator
*   **Primary Goal:** Reconcile conflicting priority signals between an OpenClaw "Security Specialist" and a Claude Code "Feature Developer."
*   **The Happy Path (Tasks):**
    1.  The user signs a "Mission Constitution" (YAML) using a local TPM-bound key.
    2.  MCP Any stores the constitution in the hardware-locked ASA registry.
    3.  Claude Code sends a "Ship Now" instruction; OpenClaw sends a "Security Audit First" instruction.
    4.  The specialist agent queries the ASA Broker via the `arbitrate` endpoint.
    5.  The ASA Broker evaluates the instructions against the Constitution (where Security > Speed) and issues a signed "Winning Intent" token.
    6.  The specialist agent executes the security audit, citing the ASA token as its authority.

## 4. Design & Architecture
*   **System Flow:**
    ```mermaid
    graph TD
        User[User/Admin] -->|Sign Constitution| TPM[Local TPM]
        TPM -->|Signed YAML| ASA[ASA Broker]
        SupA[Supervisor A: Claude] -->|Task A| Spec[Specialist Agent]
        SupB[Supervisor B: OpenClaw] -->|Task B| Spec
        Spec -->|Arbitration Request| ASA
        ASA -->|Evaluate vs Constitution| Eval[Constitutional Engine]
        Eval -->|Winning Intent Token| Spec
        Spec -->|Execute| Tool[MCP Tool]
    ```
*   **APIs / Interfaces:**
    *   `POST /v1/asa/constitution`: Register or update a signed constitution.
    *   `POST /v1/asa/arbitrate`: Request resolution for a list of conflicting instructions.
*   **Data Storage/State:**
    *   Constitutions are stored in a versioned, append-only SQLite store, with SHA-256 integrity checks against the TPM root.

## 5. Alternatives Considered
*   **Centralized Supervisor Resolution:** Rejected due to the "Supervisor Bottleneck" and single-point-of-failure risks in deep swarms.
*   **LLM-based Arbitration:** Rejected as the primary mechanism due to non-deterministic outputs and high latency. ASA uses deterministic logic for priority mapping.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** All arbitration requests must be accompanied by a hardware-attested session token.
*   **Observability:** Every arbitration event is logged to the `ASA_Audit_Log` with the winning criteria cited.

## 7. Evolutionary Changelog
*   **2026-11-02:** Initial Document Creation.
