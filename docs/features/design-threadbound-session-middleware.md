# Design Doc: Threadbound Session Middleware
**Status:** Draft
**Created:** 2026-03-10

## 1. Context and Scope
As AI agents move from single-user desktop tools to multi-user environments (Discord, Slack, shared web dashboards), the traditional stateless MCP adapter model is insufficient. Without strict session isolation, an agent's tool calls, blackboard state, or context could be accidentally (or maliciously) shared between different user conversations. This "Context Mixing" is a major security risk. MCP Any must implement "Threadbound" session isolation at the middleware layer to ensure that every agent interaction is cryptographically bound to a unique session/thread, preventing cross-user data leakage.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement mandatory `SessionID` and `ThreadID` headers for all incoming MCP requests.
    * Automatically isolate all Blackboard (Shared KV Store) reads and writes by `ThreadID`.
    * Ensure `Upstream Adapters` (HTTP, CMD, etc.) inherit the thread-specific credentials and environment variables.
    * Provide a cryptographically secure session attestation mechanism for multi-user platforms.
* **Non-Goals:**
    * Managing the user's platform-level identity (e.g., Discord Auth) – MCP Any receives a verified identity.
    * Encrypting individual tool calls (handled by the transport layer).

## 3. Critical User Journey (CUJ)
* **User Persona:** Developer of a Discord-based AI Agent Swarm.
* **Primary Goal:** Prevent User A from seeing User B's temporary research data stored in the Shared KV Store.
* **The Happy Path (Tasks):**
    1. User A sends a message in Discord.
    2. The Discord Bot attaches `ThreadID: discord-thread-A` to the MCP request.
    3. MCP Any's Threadbound Middleware intercepts the request.
    4. The agent calls `blackboard.set("search_results", "...")`.
    5. MCP Any writes to the database with `WHERE thread_id = 'discord-thread-A'`.
    6. User B sends a message in a different thread.
    7. The agent calls `blackboard.get("search_results")`.
    8. MCP Any returns `null` because User B's thread has no such key.

## 4. Design & Architecture
* **System Flow:**
    `Agent (with Thread Headers)` -> `Threadbound Middleware` -> `Blackboard / Upstream Tools`
    1. **Header Validation**: Middleware ensures the presence of a valid, signed `ThreadID` header.
    2. **Context Injection**: The `ThreadID` is injected into the Go `context.Context` for all downstream operations.
    3. **Database Isolation**: The SQL driver (used by Blackboard) uses a "Thread-Aware" query rewriter to append isolation clauses.
    4. **Credential Scoping**: If the session has specific API keys (e.g., from the `Secrets Manager`), only those are made available to the current thread.
* **APIs / Interfaces:**
    * `SetThreadSession(id, metadata)`: Sets the current thread's session context.
    * `GetThreadContext()`: Returns the metadata for the current thread.
* **Data Storage/State:**
    * `sessions.db`: SQLite table mapping `ThreadID` to session metadata, TTLs, and scoped credentials.

## 5. Alternatives Considered
* **Separate Database per User**: Rejected because it's too heavy for thousands of concurrent Discord threads.
* **Client-side isolation**: Rejected because the gateway must be the "Source of Truth" for security.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: All sessions are untrusted until verified. MFA can be required for specific threads.
* **Observability**: The `Recursive Context Dashboard` will visualize active threads and their resource usage.

## 7. Evolutionary Changelog
* **2026-03-10:** Initial Document Creation.
