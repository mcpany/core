# Design Doc: SKILL.md Interop Bridge
**Status:** Draft
**Created:** 2026-03-10

## 1. Context and Scope
The AI agent ecosystem (Claude Code, Gemini CLI, Cursor) has converged on the `SKILL.md` format for defining agent "skills" (playbooks). However, many specialized or legacy agents do not yet support this format. MCP Any will bridge this gap by dynamically converting `SKILL.md` playbooks into standard MCP tool definitions, making these skills universal.

## 2. Goals & Non-Goals
* **Goals:**
    * Automatically discover `.skill/` directories or `SKILL.md` files in a project.
    * Parse natural language instructions and templates into structured MCP Tool definitions.
    * Execute skill-based tasks by mapping MCP tool calls back to the skill's execution logic (scripts or prompts).
* **Non-Goals:**
    * Authoring `SKILL.md` files (this is a consumption/bridge feature).
    * Replacing native skill support in agents that already have it.

## 3. Critical User Journey (CUJ)
* **User Persona:** Developer with a custom Python agent.
* **Primary Goal:** Use the popular "Claude-Code-Frontend" skill within their custom agent without rewriting the skill logic.
* **The Happy Path (Tasks):**
    1. User places a `frontend.skill/SKILL.md` in their project.
    2. MCP Any detects the skill and generates an MCP tool named `skill_frontend_design`.
    3. The custom agent lists tools and sees `skill_frontend_design`.
    4. Agent calls the tool; MCP Any executes the skill's playbook logic and returns the result.

## 4. Design & Architecture
* **System Flow:**
    `Skill Scanner` -> `Playbook Parser` -> `MCP Tool Synthesizer` -> `Tool Registry`
    1. **Discovery**: Watches for `SKILL.md` or `.skill` folders.
    2. **Synthesis**: Uses a lightweight LLM (or regex parser) to map "Commands" in the skill to "Tool Schemas."
* **APIs / Interfaces:**
    * `mcp.list_tools` will now include synthesized `skill_*` tools.
* **Data Storage/State:**
    * Skill metadata is cached in the `Shared KV Store` to avoid redundant parsing.

## 5. Alternatives Considered
* **Manual Conversion**: Rejected as it defeats the purpose of "Indispensable Infrastructure."
* **Native Plugin for Agents**: Too much maintenance overhead for every agent framework.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: Skills often contain executable scripts. These must be passed through the `Project Configuration Security Guard` before activation.
* **Observability**: Track which agents are calling which skills to provide "Skill Usage Analytics."

## 7. Evolutionary Changelog
* **2026-03-10:** Initial Document Creation.
