# Design Doc: Universal SKILL.md Runtime
**Status:** Draft
**Created:** 2026-03-11

## 1. Context and Scope
With the rise of standardized `SKILL.md` files in ecosystems like Claude Code and Gemini CLI, agents are no longer just using individual tools; they are following complex, multi-step playbooks. Currently, these skills are often locked into specific agent frameworks. MCP Any needs to provide a universal runtime that can parse these `SKILL.md` files and expose their logic as high-level MCP tools, allowing any connected agent to leverage production-grade UI generation, web automation, or database optimization playbooks.

## 2. Goals & Non-Goals
* **Goals:**
    *   Implement a parser for the `SKILL.md` format (Markdown with frontmatter and executable blocks).
    *   Expose skill-defined playbooks as standard MCP tools.
    *   Ensure cross-platform compatibility (the same skill works for a Claude-based or Gemini-based agent).
    *   Support "Auto-Triggering" where MCP Any suggests a skill based on the current task context.
* **Non-Goals:**
    *   Defining a new skill format (we adhere to the existing `SKILL.md` standard).
    *   Providing the LLM logic for the skills (MCP Any handles the runtime and tool exposure).

## 3. Critical User Journey (CUJ)
* **User Persona:** Full-Stack Developer using a custom agent swarm.
* **Primary Goal:** Use a specialized `frontend-design.skill.md` playbook within an OpenClaw swarm without manual re-implementation.
* **The Happy Path (Tasks):**
    1.  User drops `frontend-design.skill.md` into the MCP Any skills directory.
    2.  MCP Any automatically parses the file and registers a new tool: `execute_skill_frontend_design`.
    3.  A subagent in the OpenClaw swarm receives a UI task and calls `execute_skill_frontend_design(task="Create a responsive navbar")`.
    4.  MCP Any executes the instructions and templates defined in the skill file, providing the structured output back to the agent.

## 4. Design & Architecture
* **System Flow:**
    - **Skill Ingestion**: A watcher service monitors the `/skills` directory for `.md` files.
    - **Parsing**: The `SkillParser` extracts the `description`, `allowed-tools`, and instruction blocks.
    - **Execution Environment**: Skills are executed in a "Detached Sandbox" to ensure they don't have unauthorized host access.
    - **MCP Mapping**: Each skill is mapped to a `call_tool` request where the skill instructions are injected into the agent's context or executed as a sub-process.
* **APIs / Interfaces:**
    - `skills/list`: Returns all available skills and their schemas.
    - `skills/execute`: The endpoint for triggering a specific skill playbook.
* **Data Storage/State:** Skill definitions are stored in memory; execution state is tracked in the `Shared KV Store` (Blackboard) for long-running playbooks.

## 5. Alternatives Considered
* **Native Agent Implementation**: Letting each agent framework handle skills. *Rejected* because it leads to fragmentation and duplicate effort for users.
* **Converting Skills to Go/Python Code**: *Rejected* because it violates the "Configuration over Code" principle of MCP Any.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Skills must explicitly declare `allowed-tools`. MCP Any enforces these boundaries via the Policy Firewall. Malicious playbooks cannot escape the sandbox.
* **Observability:** Skill execution steps are logged in the "Tool Activity Feed" in the UI, allowing users to see exactly how a skill is being processed.

## 7. Evolutionary Changelog
* **2026-03-11:** Initial Document Creation.
