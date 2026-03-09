# Design Doc: Threadbound Session Isolation
**Status:** Draft
**Created:** 2026-03-09

## 1. Context and Scope
Agents running on multi-user platforms like Discord, WhatsApp, or Slack often use a single long-running MCP Any instance. Without strict isolation, there is a risk of User A's context, secrets, or tool-state leaking to User B's session. This is especially critical given the "OpenClaw" security incidents where state leakage was observed.

## 2. Goals & Non-Goals
* **Goals:**
    * Ensure every agent interaction is bound to a cryptographically unique `ThreadID`.
    * Isolate file-system access and environment variables per thread.
    * Prevent "Cross-Talk" where an agent in one thread can access the memory of another.
* **Non-Goals:**
    * Providing a full virtual machine for every user (too heavy).
    * Replacing existing platform-specific encryption (e.g., WhatsApp's E2EE).

## 3. Critical User Journey (CUJ)
* **User Persona:** A multi-tenant bot provider running OpenClaw for 1,000 users.
* **Primary Goal:** Prevent User A from seeing User B's recently used file paths or API keys.
* **The Happy Path (Tasks):**
    1. User A sends a message via Discord.
    2. The adapter attaches a signed `X-MCP-Thread-ID` header.
    3. MCP Any creates/retrieves an isolated workspace and transient vault for this ID.
    4. Tool calls made by this agent are scoped to this workspace.
    5. User B sends a message; MCP Any ensures User B cannot access User A's scoped resources.

## 4. Design & Architecture
* **System Flow:**
    * **Middleware Layer**: Validates the `ThreadID` signature.
    * **Scoped Context Provider**: Intercepts tool calls and injects thread-specific paths or credentials.
    * **Transient Vault**: In-memory encrypted storage for secrets that expires with the session.
* **APIs / Interfaces:**
    * New Internal Interface: `IsolationManager.GetScope(threadID string) -> Scope`
* **Data Storage/State:**
    * Scoped SQLite databases for long-term thread state.
    * Encrypted memory-mapped files for transient secrets.

## 5. Alternatives Considered
* **Process Isolation (Docker)**: Rejected for low-latency chat apps due to overhead, though could be a P2 option for high-security tiers.
* **Logical Isolation (Namespacing)**: Currently used but proven insufficient in recent "OpenClaw" exploits.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** If a `ThreadID` is compromised, the blast radius is limited to that single user's context.
* **Observability:** Trace all tool calls with the associated `ThreadID`.

## 7. Evolutionary Changelog
* **2026-03-09:** Initial Document Creation.
