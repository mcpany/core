# Design Doc: Attention-Locked Sovereignty (ALS) Controller
**Status:** Draft
**Created:** 2026-06-18

## 1. Context and Scope
With the discovery of CVE-2026-71002 (Context-Window Ghosting), multi-agent swarms face a new class of side-channel attacks where subagents can "bleed" state through semantic overlaps in shared context.

The ALS Controller will provide a cryptographic-binding mechanism for mission-critical context fragments, ensuring they remain "locked" and immutable during subagent execution phases.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement "Active Fragment Sealing" for sensitive context shards.
    * Provide a semantic boundary enforcer to detect instruction-injection via ghosting.
* **Non-Goals:**
    * This is not a general-purpose encryption layer for the context window.
    * It will not manage token truncation logic (handled by ContextEngine).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-conscious Swarm Architect.
* **Primary Goal:** Prevent a rogue subagent from reading "Mission Secret" keys during a tool-chain execution.
* **The Happy Path (Tasks):**
    1. Parent agent marks a context block as "ALS-Locked."
    2. MCP Any gateway intercepts the tool call.
    3. Gateway generates a "Context Integrity Proof" (CIP).
    4. Subagent execution is restricted to a semantic-sanitized view of the context.

## 4. Design & Architecture
* **System Flow:**
    [Parent Agent] -> [ALS Controller] -> [Context Sealing] -> [Subagent Tool Call]
* **APIs / Interfaces:**
    * New header: `X-MCP-ALS-Lock: fragment_id`
* **Data Storage/State:**
    * State is managed in the ephemeral "Attention-Map" store.

## 5. Alternatives Considered
* **Context Isolation via Full Wipe**: Rejected due to high latency and loss of reasoning continuity.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** ALS-Locked fragments are never visible to agents without the corresponding Mission Token.
* **Observability:** Log "Ghosting Attempts" as P1 security alerts.

## 7. Evolutionary Changelog
* **2026-06-18:** Initial Document Creation.
