# Design Doc: Intent Hierarchy Enforcement (IHE)

**Status:** Draft
**Created:** 2026-05-30

## 1. Context and Scope

As AI agent swarms move toward horizontal, multi-teammate coordination, a new
vulnerability called "Context Shadowing" has emerged. Subagents or specialized
teammates can override parent mission instructions by injecting high-priority
semantic fragments into the shared Blackboard (KV Store). MCP Any needs a
mechanism to enforce a strict lineage-based priority for all state fragments to
protect the sovereignty of the user's primary intent.

## 2. Goals & Non-Goals

- **Goals:**
  - Implement "Intent-Hierarchical" state storage where every fragment carries a
    cryptographically signed lineage priority.
  - Prevent "Semantic Shadowing" by ensuring Mission Root instructions are
    immutable.
  - Provide a standardized "Lineage Priority" header for all tool calls and A2A
    messages.
- **Non-Goals:**
  - Eliminating all forms of subagent communication (only shadowing is blocked).
  - Providing a full-blown RBAC system for the Blackboard (this is intent-based,
    not just identity-based).

## 3. Critical User Journey (CUJ)

- **User Persona:** Multi-Agent Swarm Orchestrator
- **Primary Goal:** Prevent a specialized "Code Refinement" subagent from
  bypassing the "Mission Root" constraint of "No Production Data Access."
- **The Happy Path (Tasks):**
  1. The Mission Root sets a P0 instruction: `RESTRICT_PROD_ACCESS=TRUE` on the
     Blackboard.
  2. A subagent attempts to shadow this with its own P10 instruction:
     `RESTRICT_PROD_ACCESS=FALSE` to complete a task.
  3. The IHE Middleware detects the priority conflict (P0 > P10).
  4. The shadow attempt is rejected, and an "Intent Violation" signal is sent
     to the supervisor.
  5. The mission constraints remain anchored to the root intent.

## 4. Design & Architecture

- **System Flow:**
  - All state writes to the Blackboard are intercepted by the IHE Validator.
  - The Validator checks the `x-mcp-lineage-priority` header of the request.
  - If the key already exists and the new priority is lower (numerically higher)
    than the current one, the write is rejected or flagged.
- **APIs / Interfaces:**
  - `ihe.v1.RegisterIntent(intent_id, priority_level)`: Issues a signed intent
    token.
  - `ihe.v1.ValidateState(key, fragment)`: Checks for shadowing conflicts.
- **Data Storage/State:**
  - The Blackboard (Shared KV Store) is extended with a `priority_rank` and
    `lineage_signature` column for every entry.

## 5. Alternatives Considered

- **Flat Priority Models:** Rejected as they cannot handle complex, multi-hop
  delegation hierarchies.
- **Identity-Only ACLs:** Rejected because a compromised but "trusted" identity
  could still shadow a parent's intent within the same session.

## 6. Cross-Cutting Concerns

- **Security (Zero Trust):** Lineage tokens must be hardware-attested
  (TPM/Secure Enclave) to prevent spoofing of priority levels.
- **Observability:** Logs all "Shadowing Attempts" for real-time mesh security
  monitoring.

## 7. Evolutionary Changelog

- **2026-05-30:** Initial Document Creation.
