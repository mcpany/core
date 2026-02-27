# Design Doc: Strict Seatbelt Policies
**Status:** Draft
**Created:** 2026-02-27

## 1. Context and Scope
As agents gain more autonomous capabilities (e.g., file system access, shell execution), the risk of catastrophic accidental or malicious actions increases. Gemini CLI (v0.30.0) introduced "Strict Seatbelt Profiles" to enforce rigid safety boundaries. MCP Any needs a universal implementation of this concept that can be applied across any agent framework and any MCP server.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide "Seatbelt Profiles" (e.g., `read-only`, `local-only`, `sandbox-heavy`) that can be applied to sessions.
    * Enforce rigid safety boundaries at the middleware layer before tool execution.
    * Allow dynamic hardening of seatbelts based on the task phase.
    * Support grounded "Infrastructure-Grade" auditing of tool upstreams.
* **Non-Goals:**
    * Replacing the main Policy Firewall (Seatbelts are a "Quick-Hardening" layer on top).
    * Providing an interactive UI for every single tool call (handled by HITL middleware).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Developer
* **Primary Goal:** Run an experimental community "Claw Skill" without risking host-level file deletion.
* **The Happy Path (Tasks):**
    1. User starts MCP Any with `--policy seatbelt:read-only`.
    2. User enables a new OpenClaw skill that connects via MCP Any.
    3. The skill attempts to call `fs:delete` on a local file.
    4. The Seatbelt Middleware intercepts the call, identifies it as a "write" operation blocked by the `read-only` profile.
    5. The call is blocked, and a security audit alert is generated.

## 4. Design & Architecture
* **System Flow:**
    `Agent -> Session Manager -> Seatbelt Middleware (Profile Lookup) -> Policy Engine -> Tool Execution`
* **APIs / Interfaces:**
    * `--policy seatbelt:<profile_name>`: CLI flag to set global or session-bound seatbelts.
    * `X-MCP-Seatbelt-Profile`: Header to propagate seatbelt requirements in multi-agent handoffs.
* **Data Storage/State:**
    * Seatbelt profiles are defined in `config.yaml` or as internal presets.

## 5. Alternatives Considered
* **Granular RBAC**: Rejected for this specific use case because it's too complex to configure for "quick hardening." Seatbelts provide coarse-grained, high-confidence safety.
* **Agent-Side Enforcement**: Rejected because agents cannot be trusted to enforce their own safety boundaries in "Zero Trust" environments.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: Seatbelts are the "Last Line of Defense" and are applied regardless of the tool's own permissions.
* **Observability**: All "Seatbelt Interceptions" are logged with high severity and include the full `SessionContext`.

## 7. Evolutionary Changelog
* **2026-02-27:** Initial Document Creation (Inspired by Gemini CLI v0.30.0).
