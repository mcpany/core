# Design Doc: Deterministic Permission Guard (DPG)
**Status:** Draft
**Created:** 2026-05-09

## 1. Context and Scope
The "Permission Bypass" failures (e.g., Claude Code Bug #8961) demonstrate that instruction-based security is a flawed paradigm. Agents can hallucinate or be coerced into ignoring "Deny" rules in their prompt. The Deterministic Permission Guard (DPG) shifts security from the reasoning layer to the infrastructure layer, enforcing project-local "Deny" rules at the middleware level.

## 2. Goals & Non-Goals
* **Goals:**
    * Enforce non-bypassable "Deny" rules for tool calls and filesystem access.
    * Operate independently of the agent's internal reasoning or prompt state.
    * Integrate with kernel-level Inode pinning to prevent symlink-racing escapes.
* **Non-Goals:**
    * Replacing the high-level Policy Firewall (Rego/CEL).
    * Dynamic "Allow" rule generation (DPG is primarily for static "Deny" enforcement).

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise Security Administrator
* **Primary Goal:** Ensure an agent never accesses `.env` files or production credentials, even if it's compromised.
* **The Happy Path (Tasks):**
    1. Admin defines a DPG policy in `.mcpany/security.yaml` with a `deny` rule for `**/.env`.
    2. User initiates an agent session.
    3. Agent (compromised or hallucinating) attempts to call `fs:read` on `.env`.
    4. The DPG middleware intercepts the tool call.
    5. DPG identifies the target path matches a "Deny" rule.
    6. DPG immediately rejects the call with a `SECURITY_BLOCK` error, before it reaches the tool executor.

## 4. Design & Architecture
* **System Flow:**
    - `Agent` -> `ToolCall(path)` -> `DPG Middleware`
    - `DPG Middleware` -> `Match(path, DenyRules)` -> `Action(Block/Pass)`
* **APIs / Interfaces:**
    - Internal middleware hook in the MCP server request pipeline.
    - Configuration schema for `.mcpany/security.yaml`.
* **Data Storage/State:**
    - Rules are loaded into an optimized prefix tree (Trie) for O(1) matching.

## 5. Alternatives Considered
* **Prompt-Based Deny Rules:** Rejected as proven bypassable (Bug #8961).
* **OS-Level Permissions:** Rejected due to lack of granularity and difficulty in cross-platform synchronization.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** DPG is the "Last Line of Defense," operating after all other policy checks.
* **Observability:** Blocked attempts are flagged in the Local Security Audit Log with high severity.

## 7. Evolutionary Changelog
* **2026-05-09:** Initial Document Creation.
