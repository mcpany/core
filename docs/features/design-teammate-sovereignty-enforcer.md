# Design Doc: Teammate Sovereignty Enforcer (TSE)
**Status:** Draft
**Created:** 2026-06-18

## 1. Context and Scope
With the rise of "Agent Teams" (Claude Code), multiple subagents now run in parallel.

## 2. Goals & Non-Goals
* **Goals:** Cryptographically isolate session state between parallel subagents.

## 3. Critical User Journey (CUJ)
1. Orchestrator spawns 3 agents.
2. TSE assigns unique UID/Namespace to each.

## 4. Design & Architecture
- **System Flow:** Uses Docker-bound named pipes.

## 5. Alternatives Considered
- Shared filesystem.

## 6. Cross-Cutting Concerns
- **Security:** Zero Trust.

## 7. Evolutionary Changelog
- **2026-06-18:** Initial Document Creation.
