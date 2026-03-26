# Design Doc: Teammate Sovereignty Enforcer (TSE)
**Status:** Draft
**Created:** 2026-06-18

## 1. Context and Scope
With the rise of "Agent Teams" (Claude Code), multiple subagents now run in parallel. A compromise in one subagent can lead to "State Smearing" across the entire team.

## 2. Goals & Non-Goals
* **Goals:** Cryptographically isolate session state between parallel subagents.
* **Non-Goals:** Does not manage LLM token budgets.

## 3. Critical User Journey (CUJ)
1. Orchestrator spawns 3 agents.
2. TSE assigns unique UID/Namespace to each.
3. Agent A cannot read Agent B's /tmp or env vars.

## 4. Design & Architecture
- **System Flow:** Uses Docker-bound named pipes and Linux Namespaces.
- **APIs:** `/v1/session/isolate`

## 5. Alternatives Considered
- Shared filesystem with ACLs (Rejected: Too easy to misconfigure).

## 6. Cross-Cutting Concerns
- **Security:** Zero Trust at the pipe level.

## 7. Evolutionary Changelog
- **2026-06-18:** Initial Document Creation.
