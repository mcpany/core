# Design Doc: Swarm Governance & Cross-Skill Orchestration

**Status:** Draft
**Created:** 2026-03-11

## 1. Context and Scope
The AI agent landscape is rapidly shifting from single-agent sequential tasks to "Agent Teams" and "Swarms" (e.g., Claude Agent Teams, OpenClaw swarms). This parallelism introduces new challenges: "Agent Storms" (DDoS-like surges in tool requests), state synchronization stalls, and the need for standardized agent "playbooks" (Skills). MCP Any must evolve to govern these swarms, ensuring stability and providing a universal bridge for standardized Skills.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Implement **Swarm-Aware Rate Limiting** to prevent resource exhaustion from parallel agent calls.
    *   Support the **Universal Skill Adapter (`.SKILL.md`)** for automatic tool and policy configuration.
    *   Upgrade the **Shared KV Store (Blackboard)** for parallel-safe, lock-free concurrency.
    *   Provide **Swarm Quota Enforcement** via the Policy Firewall.
*   **Non-Goals:**
    *   Orchestrating the agents' internal logic or teammate selection.
    *   Developing new agent models or LLMs.

## 3. Critical User Journey (CUJ)
*   **User Persona:** Enterprise Swarm Architect
*   **Primary Goal:** Deploy a parallel team of 10 code-review agents that share state securely and stay within rate limits.
*   **The Happy Path (Tasks):**
    1.  Architect defines a `.SKILL.md` playbook for "Parallel Code Review."
    2.  MCP Any ingests the Skill, automatically configuring the necessary GitHub MCP server and enabling a "Reviewer Swarm" quota.
    3.  10 agents initialize in parallel, all inheriting the same "Swarm Session ID."
    4.  As agents call tools, MCP Any's Swarm-Aware Rate Limiter ensures they don't exceed the GitHub API limits.
    5.  Agents write review notes to the Blackboard simultaneously; MCP Any uses optimistic merging to ensure no state is lost.

## 4. Design & Architecture
*   **System Flow:**
    - **Skill Ingestion**: The `SkillAdapter` parses `.SKILL.md`, extracts tool requirements, and generates a dynamic MCP configuration profile.
    - **Rate Limiting Middleware**: A token-bucket algorithm scoped to the `Swarm-ID`. It intercepts all `tools/call` requests.
    - **Blackboard (CRDT/Optimistic)**: Shifting from standard SQLite transactions to an optimistic concurrency model (or row-level versioning) to allow non-blocking writes from parallel agents.
*   **APIs / Interfaces:**
    - `POST /skills/ingest`: Upload and activate a `.SKILL.md` file.
    - `GET /swarm/{swarm_id}/metrics`: View real-time quota usage and concurrency stats.
*   **Data Storage/State:**
    - Swarm quotas and active token counts are stored in-memory (Redis-backed for distributed deployments).
    - Persistent state remains in the isolated Blackboard.

## 5. Alternatives Considered
*   **Wait-and-Retry Logic in Agents**: Rejected because it leads to "Thundering Herd" problems and wastes LLM tokens. Centralized governance in MCP Any is more efficient.
*   **Manual Skill Mapping**: Rejected because the `.SKILL.md` standard is becoming the industry "Playbook" standard; manual mapping doesn't scale.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** Skills must be attested before ingestion. Quotas prevent a single compromised swarm from "burning" a team's entire API budget.
*   **Observability:** The UI will feature a "Swarm Pulse" view, visualizing parallel calls and rate-limiting "throttling" events.

## 7. Evolutionary Changelog
*   **2026-03-11:** Initial Document Creation.
