# Design Doc: Cognitive Proof-of-Work (CPoW) Gateway
**Status:** Draft
**Created:** 2026-06-18

## 1. Context and Scope
To protect against machine-speed coercion of sensitive tools (e.g., direct shell access), MCP Any needs a secondary validation layer that requires a specialized subagent or user to provide a "Cognitive Proof" that the action aligns with the original mission.

## 2. Goals & Non-Goals
* **Goals:**
    * Force high-risk tools to wait for an asynchronous "Confirmation Fragment."
    * Require multi-agent signatures for specific capability-cards.
* **Non-Goals:**
    * Replace existing HITL workflows (it augments them).
    * Provide a general-purpose bidding bus (handled by the DCA Auction Broker).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Prevent an autonomous subagent from accidentally deleting the `.git/` directory.
* **The Happy Path (Tasks):**
    1. Agent calls `rm -rf .git`.
    2. CPoW Gateway intercepts the call.
    3. Hub requests a "Cognitive Justification" from a separate "Security Auditor" agent.
    4. Action is committed only if both agents provide matching reasoning fragments.

## 4. Design & Architecture
* **System Flow:** `Agent (Action)` -> `CPoW Middleware (Wait)` -> `Auditor Agent (Justification)` -> `CPoW Gateway (Merge)` -> `Execute`.
* **APIs / Interfaces:** `RequestProof(ActionID)`, `SubmitProof(Justification)`.
* **Data Storage/State:** Ephemeral "Waiting Room" for pending high-risk calls.

## 5. Alternatives Considered
* **Always-on HITL:** Rejected because it causes "Approval Fatigue" for low-risk actions. CPoW is selectively applied to high-risk tools.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Mitigates "Prompt Injection" as a second-order validator.
* **Observability:** Logs all CPoW-rejected actions for offline behavioral analysis.

## 7. Evolutionary Changelog
* **2026-06-18:** Initial Document Creation.
