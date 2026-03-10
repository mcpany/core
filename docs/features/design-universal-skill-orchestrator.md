# Design Doc: Universal Skill Orchestrator
**Status:** Draft
**Created:** 2026-03-10

## 1. Context and Scope
The agent ecosystem is converging on the `SKILL.md` format for defining specialized agent playbooks and templates. Leading agents like Claude Code, Cursor, and Gemini CLI use these files to extend their capabilities beyond simple tool calls. However, these skills are often siloed within specific agent frameworks. MCP Any, as a universal adapter, must bridge this gap by ingesting `SKILL.md` files and exposing them as high-level "Super-Tools" via the standard MCP protocol. This allows any MCP-compatible agent to benefit from project-local skills without needing native `SKILL.md` support.

## 2. Goals & Non-Goals
* **Goals:**
    * Automatically discover and parse `SKILL.md` files in project directories.
    * Transform skill definitions (templates, instructions, constraints) into MCP Tool schemas.
    * Orchestrate the execution of multiple underlying MCP tools to fulfill a skill's intent.
    * Provide a standardized "Skill Context" that agents can query on-demand.
* **Non-Goals:**
    * Replacing the natural language reasoning of the agent (the agent still decides *when* to use a skill).
    * Modifying the `SKILL.md` format itself (must remain compatible with the emerging standard).

## 3. Critical User Journey (CUJ)
* **User Persona:** Agent Developer using a framework that doesn't natively support `SKILL.md`.
* **Primary Goal:** Use a repository's specialized `/frontend-review` skill via an MCP-only agent.
* **The Happy Path (Tasks):**
    1. MCP Any detects `SKILL.md` in the project root.
    2. MCP Any exposes a new tool: `skill_frontend_review`.
    3. The agent calls `skill_frontend_review` with the target file path.
    4. MCP Any's Skill Orchestrator reads the skill's instructions and template.
    5. MCP Any coordinates with the `FileSystem` and `Linter` tools to gather data.
    6. MCP Any returns the formatted skill output to the agent.

## 4. Design & Architecture
* **System Flow:**
    `Agent` -> `MCP Any Server` -> `Skill Orchestrator Adapter` -> `SKILL.md Files`
    1. **Discovery**: A background watcher identifies `.md` files matching the `SKILL.md` pattern or containing skill metadata.
    2. **Parsing**: A parser converts Markdown sections (Context, Instructions, Tools) into a structured `SkillDefinition` object.
    3. **Mapping**: The `SkillDefinition` is mapped to an MCP `Tool` schema, where sections are exposed as parameters or tool descriptions.
    4. **Execution**: When called, the adapter acts as a "mini-agent" or template-engine, resolving references to other MCP tools.
* **APIs / Interfaces:**
    * `SkillAdapter.ListTools()`: Returns all discovered skills as MCP tools.
    * `SkillAdapter.CallTool(name, args)`: Executes the skill logic.
* **Data Storage/State:**
    * `SkillRegistry`: In-memory cache of parsed skills with file-system watchers for hot-reloading.

## 5. Alternatives Considered
* **Client-side Skill Parsing**: Rejected because it requires every agent client to implement the parser.
* **Static Config mapping**: Rejected because `SKILL.md` is dynamic and project-specific; manual mapping would be too high-friction.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: Skills are subjected to the `Policy Firewall`. Execution of composite tools is restricted by the agent's current `Intent Scope`.
* **Observability**: Skill execution is traced as a parent span, with individual tool calls as child spans in the `Tool Activity Feed`.

## 7. Evolutionary Changelog
* **2026-03-10:** Initial Document Creation.
