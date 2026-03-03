# Design Doc: Semantic Path Scoping & Intent Verification
**Status:** Draft
**Created:** 2026-03-02

## 1. Context and Scope
The recent CVE-2026-27735 (Path Traversal in `mcp-server-git`) and CVE-2026-1977 (Code Injection in `mcp-vegalite-server`) highlight that simple filesystem permissions are insufficient for autonomous agents. Agents often need access to a "Project Root" but should not be allowed to escape it or perform "Out-of-Scope" modifications even if the tool allows it. MCP Any needs a layer that verifies tool arguments against both a physical sandbox and a semantic intent.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement "Path-Bound Scoping" middleware that enforces a strict boundary for all file-related tool calls.
    * Introduce "Intent Verification" where tools with side effects (Write, Delete, Execute) must match a high-level goal declared in the session context.
    * Provide a "Semantic Jail" for tools like Git, Shell, and Filesystem to prevent traversal and unauthorized execution.
* **Non-Goals:**
    * Replacing OS-level sandboxing (e.g., Docker).
    * Writing specific logic for every possible tool (must be a generic middleware).

## 3. Critical User Journey (CUJ)
* **User Persona:** Developer using an Agent Swarm to refactor a legacy codebase.
* **Primary Goal:** Ensure the agent only modifies files within `/home/user/project` and doesn't accidentally run `git add /etc/passwd`.
* **The Happy Path (Tasks):**
    1. User starts MCP Any with a session scope restricted to `/home/user/project`.
    2. Agent attempts a `git add ../secret.txt`.
    3. Semantic Middleware intercepts the call, detects the traversal, and blocks it before it reaches the Git tool.
    4. User receives an alert: "Blocked: Path outside of session scope."

## 4. Design & Architecture
* **System Flow:**
    - **Argument Interception**: Middleware hooks into the tool execution pipeline.
    - **Path Resolution**: All string arguments that look like paths are resolved to absolute paths.
    - **Scope Check**: Resolved paths are checked against the `SessionScope` allowed roots.
    - **Semantic Analysis**: (Optional) For P0, simple regex/glob matching. For future, LLM-based intent verification.
* **APIs / Interfaces:**
    - `PolicyEngine.Verify(tool string, args map[string]any, scope SessionScope) error`
* **Data Storage/State:** `SessionScope` is stored in the active agent session state.

## 5. Alternatives Considered
* **Individual Tool Hardening**: Fixing every tool one by one. *Rejected* as it's unscalable (the "Whack-a-Mole" problem).
* **Docker-Only Execution**: Running every tool in its own container. *Rejected* due to extreme latency and complexity for local developers.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Critical for preventing agents from becoming "Internal Threat Actors."
* **Observability:** Every blocked call must be logged with the "Reason for Denial" and the specific argument that triggered it.

## 7. Evolutionary Changelog
* **2026-03-02:** Initial Document Creation.
