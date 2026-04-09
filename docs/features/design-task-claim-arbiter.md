# Design Doc: Task-Claim Arbiter (TCA)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
With the maturation of horizontal "Agent Teams" (e.g., Claude Code, CrewAI), agents now coordinate via shared state (Task Boards) rather than linear hierarchies. Market analysis reveals a critical vulnerability: **Task-Claim Hijacking**. A compromised subagent or a rogue specialist can "claim" a high-privilege task from the board before the team lead can assign it to the intended verified specialist, leading to unauthorized tool execution and data exfiltration.

MCP Any must move from being a simple tool-call gateway to a **Mission Governance Hub**. The TCA will act as the authoritative host for the team's "Shared Task Board," enforcing identity and intent verification at the point of task assignment.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide a hardware-attested "Shared Task Board" for agent teams.
    * Mandate identity verification (FSI/TPM) before any agent can claim or modify a task.
    * Enforce "Lead-Only Assignment" for high-risk tasks.
    * Maintain a cryptographically signed audit trail of task ownership changes.
* **Non-Goals:**
    * Managing the internal reasoning logic of the agents.
    * Replacing existing project management tools (Jira/Linear). This is for runtime agent coordination only.

## 3. Critical User Journey (CUJ)
* **User Persona:** Swarm Orchestrator (DevOps/Security)
* **Primary Goal:** Prevent a specialized "CSS Agent" from claiming a "Database Migration" task within a parallel team.
* **The Happy Path (Tasks):**
    1. The Team Lead agent posts 3 tasks to the MCP Any Task Board.
    2. The "CSS Agent" attempts to claim the "Database Migration" task.
    3. The TCA checks the "CSS Agent's" hardware-attested **Capability Card**.
    4. The TCA identifies a mismatch between the agent's role and the task requirements.
    5. The TCA rejects the claim and alerts the Team Lead.
    6. The "Database Specialist" claims the task; the TCA verifies the identity and grants a time-bound **Task Lease**.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    sequenceDiagram
        AgentTeamLead->>TCA: Post Task (Requires: DB_ADMIN)
        SubAgent->>TCA: Claim Task (Identity: CSS_SPECIALIST)
        TCA->>IdentityVault: Verify Capability(CSS_SPECIALIST, DB_ADMIN)
        IdentityVault-->>TCA: Denied
        TCA-->>SubAgent: 403 Forbidden (Inadequate Capability)
        VerifiedSpecialist->>TCA: Claim Task (Identity: DB_EXPERT)
        TCA->>IdentityVault: Verify Capability(DB_EXPERT, DB_ADMIN)
        IdentityVault-->>TCA: Approved
        TCA-->>VerifiedSpecialist: 200 OK (Task Lease Issued)
    ```
* **APIs / Interfaces:**
    * `POST /v1/board/tasks`: Post a new task with required capabilities.
    * `PUT /v1/board/tasks/{id}/claim`: Claim a task (requires hardware-attested signature).
    * `GET /v1/board/status`: Real-time view of task ownership.
* **Data Storage/State:**
    * Tasks are stored in a kernel-resident, encrypted SQLite shard.
    * Ownership is governed by **Conflict-Free Replicated Data Types (CRDTs)** to ensure non-blocking performance in high-frequency meshes.

## 5. Alternatives Considered
* **Implicit Trust in Team Lead:** Rejected because the Lead session itself can be hijacked or coerced via indirect prompt injection.
* **Static Config-based Assignment:** Rejected because modern swarms are dynamic and require "Auction-style" bidding for tasks (UACO protocol).

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All claims must be signed using the agent's unique TPM-bound key.
* **Observability:** Real-time "Task Heatmap" in the UI Dashboard showing claim attempts and denials.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
