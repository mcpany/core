# Design Doc: Universal Skill Adapter (`SKILL.md`)
**Status:** Draft
**Created:** 2026-03-11

## 1. Context and Scope
The AI ecosystem has converged on the `SKILL.md` format as a universal way to define agent playbooks, instructions, and tool requirements. Currently, agents like Claude Code or Gemini CLI read these files directly from the local filesystem. This creates a security risk (unvalidated instructions) and a fragmentation problem (different agents supporting different subsets of the format). MCP Any should provide a "Universal Skill Adapter" that ingests `SKILL.md` files, validates them, and exposes them as first-class MCP tools and resources.

## 2. Goals & Non-Goals
* **Goals:**
    * Native parsing and validation of `SKILL.md` files.
    * Exposing skills as MCP "Prompts" and "Tools" for consumption by any agent.
    * Centralized governance and audit logging for all skill invocations.
    * Automatic "Skill Refinement": Suggesting improvements to `SKILL.md` files based on execution history.
* **Non-Goals:**
    * Inventing a new skill format (we will adhere to the community standard).
    * Modifying the agent's internal prompt engine (we act as a provider).

## 3. Critical User Journey (CUJ)
* **User Persona:** Agent Architect or Power User.
* **Primary Goal:** Share a specialized "Security Auditor" skill across both Claude Code and a local OpenClaw swarm.
* **The Happy Path (Tasks):**
    1. User drops `security-auditor.SKILL.md` into the project root.
    2. MCP Any detects the new file and parses its requirements.
    3. MCP Any verifies that all tools required by the skill are available and authorized.
    4. The skill appears as a new prompt in the MCP Any UI and is served to connected agents.
    5. User invokes the skill via their agent (e.g., `/security-auditor`).

## 4. Design & Architecture
* **System Flow:**
    `SKILL.md File` -> `Skill Parser` -> `MCP Prompt/Tool Registry` -> `AI Agent`
    1. **Discovery**: A new `FileWatcher` specifically looks for `.SKILL.md` files.
    2. **Transformation**: The parser converts `SKILL.md` sections (Instructions, Tools, Examples) into MCP-compliant Tool and Prompt definitions.
    3. **Enforcement**: The `Policy Engine` ensures that the skill doesn't request unauthorized tools.
* **APIs / Interfaces:**
    * `Prompts.list`: Includes parsed skills as available prompts.
    * `Tools.call`: Used to trigger any executable logic defined within a skill's examples or hooks.
* **Data Storage/State:**
    * `skills.db`: SQLite table tracking loaded skills, their versions, and their provenance.

## 5. Alternatives Considered
* **Manual Conversion**: Too slow and error-prone for users.
* **Agent-Specific Plugins**: Doesn't solve the "Universal" goal.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: Skills are subjected to the same provenance checks as MCP servers. Malicious instructions (prompt injection) within a skill are flagged during parsing.
* **Observability**: Skill usage is tracked in the `Audit Log`, allowing users to see exactly which agents are using which skills.

## 7. Evolutionary Changelog
* **2026-03-11:** Initial Document Creation.
