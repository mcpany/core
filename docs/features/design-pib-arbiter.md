# Design Doc: Priority-Inheriting Blackboard (PIB) Arbiter
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
In high-density horizontal swarms, specialist agents often lock Blackboard keys to perform complex refinement. However, if a high-priority Mission Root correction (e.g., "STOP" or "REPLAN") is issued, it can be blocked by these specialist locks, leading to "Negotiation Deadlock" or "Cognitive Stall." The PIB Arbiter ensures that mission-critical intents can bypass or inherit locks from lower-priority tasks.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a priority-aware locking mechanism for the Shared KV Store.
    * Enable "Priority Inheritance" where a locked resource inherits the priority of the highest-priority waiter.
    * Provide automated deadlock detection and resolution based on Mission-Root authority.
* **Non-Goals:**
    * Replacing SQLite as the underlying storage (PIB is a middleware layer).
    * Real-time guarantees for non-mission-critical tasks.

## 3. Critical User Journey (CUJ)
* **User Persona:** Real-time DevOps Response Swarm
* **Primary Goal:** Ensure a "Safety Shutdown" command reaches all teammates even if they are mid-lock on a deployment script.
* **The Happy Path (Tasks):**
    1. Specialist Agent A locks `k8s.deploy.status` to update a manifest.
    2. Mission Root issues a `ROLLBACK` intent with Priority: Emergency.
    3. The PIB Arbiter detects the conflict on `k8s.deploy.status`.
    4. Specialist Agent A's lock is "upgraded" to Emergency priority, forcing its execution to complete or be pre-empted.
    5. The Mission Root intent is processed immediately after the atomic upgrade, breaking the stall.

## 4. Design & Architecture
* **System Flow:**
    [Tool Call] -> [PIB Middleware] -> [Lock Manager (Priority-Aware)] -> [Blackboard (SQLite)]
* **APIs / Interfaces:**
    * `AcquireLock(key, priority_token)`
    * `ReleaseLock(key)`
* **Data Storage/State:**
    * Lock table tracking `key`, `owner_agent_id`, `current_priority`, and `wait_queue`.

## 5. Alternatives Considered
* **Lock Expiration (Timeouts):** Rejected because timeouts lead to inconsistent state; priority inheritance preserves atomicity.
* **Optimistic Concurrency:** Rejected for high-risk operations where state collisions must be prevented before write.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Priority tokens must be cryptographically bound to the Mission Root.
* **Observability:** Track "Lock Contention" metrics in the UI to identify inefficient swarms.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
