# Design Doc: Autonomous NHI Identity Rotation (ANIR)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
"Identity Squatting" has emerged as a major vulnerability in autonomous swarms. When a specialist agent (Non-Human Identity, or NHI) is spawned for a task, its session tokens often persist for the duration of the entire mission, even after its specific task is complete. This allows a compromised subagent to "squat" on its identity and perform unauthorized actions.

ANIR addresses this by implementing hardware-locked, task-triggered identity rotation. Credentials for NHIs are automatically rotated or revoked upon task completion or chapter transition, enforcing the principle of temporal least privilege.

## 2. Goals & Non-Goals
* **Goals:**
    * Automatically rotate NHI identity tokens upon task completion.
    * Bind identity rotation to the completion signal from the `Teammate Task-List Arbiter`.
    * Ensure newly rotated tokens have a subset of the previous token's scopes (Privilege-Constrained).
    * Utilize TPM-bound monotonic counters to prevent token replay during rotation.
* **Non-Goals:**
    * Managing human user identity rotation.
    * Rotating identities for third-party external services (out of scope for local NHI governance).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Prevent a "Code Auditor" subagent from using its credentials to edit files after it has finished auditing.
* **The Happy Path (Tasks):**
    1. A subagent is spawned with an `audit` identity to perform a specific code review task.
    2. The subagent finishes the task and sends a completion signal to the shared mailbox.
    3. The `Autonomous NHI Identity Rotation` service intercepts the signal.
    4. It immediately revokes the current `audit` token and issues a hardware-locked rotation receipt.
    5. If the subagent needs to perform a follow-up task, it must present the receipt to receive a new, constrained token.
    6. Any attempts to use the old `audit` token are interdicted.

## 4. Design & Architecture
* **System Flow:**
    * The ANIR service monitors the `Shared Task List` (Blackboard/Mailbox).
    * It is integrated with the `NHI Lifecycle Governance Provider`.
* **APIs / Interfaces:**
    * Internal hook: `OnTaskComplete(task_id, agent_id)` triggers rotation.
    * Internal hook: `OnChapterTransition(new_chapter_id)` triggers global NHI rotation for the session.
* **Data Storage/State:**
    * Rotation state and monotonic counters are stored in the hardware-locked `Identity Vault`.

## 5. Alternatives Considered
* **Short-lived Tokens (TTL only):** Rejected because they don't account for the speed of agent actions. A token could still be abused within its 5-minute window.
* **Manual Revocation:** Rejected as it doesn't scale for autonomous swarms.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Uses `Privilege-Constrained Token Rotation (PCTR)` to ensure no authority escalation.
* **Observability:** Rotation events are visualized in the `NHI Lifecycle Dashboard`.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
