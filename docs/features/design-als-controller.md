# Design Doc: Attention-Locked Sovereignty (ALS) Controller
**Status:** Draft
**Created:** 2026-06-18

## 1. Context and Scope
With the discovery of CVE-2026-71002 (Context-Window Ghosting), multi-agent swarms face a new class of side-channel attacks where subagents can "bleed" state through semantic overlaps in shared context.
The ALS Controller will provide a cryptographic-binding mechanism for mission-critical context fragments.

## 2. Goals & Non-Goals
* **Goals:** Active Fragment Sealing, Semantic Boundary Enforcement.
* **Non-Goals:** General context encryption, Token truncation.

## 3. Critical User Journey (CUJ)
1. Parent agent marks context as "ALS-Locked". 2. MCP Any intercepts. 3. Gateway generates CIP. 4. Subagent execution is restricted.

## 4. Design & Architecture
[Parent Agent] -> [ALS Controller] -> [Context Sealing] -> [Subagent Tool Call]

## 5. Alternatives Considered
Context Isolation via Full Wipe (rejected: high latency).

## 6. Cross-Cutting Concerns
Zero Trust security, Log ghosting attempts as alerts.

## 7. Evolutionary Changelog
* **2026-06-18:** Initial Document Creation.
