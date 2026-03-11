# Design Doc: Universal Skill Adapter
**Status:** Draft
**Created:** 2026-03-11

## 1. Context and Scope
With the rise of Claude Code and Gemini CLI, the `SKILL.md` format has become the standard for defining agent playbooks, instructions, and specialized task templates. Currently, these skills are often locked into specific IDEs or local environments. **MCP Any** needs a **Universal Skill Adapter** to ingest these `SKILL.md` files and expose them as standardized MCP tools and prompts. This allows a central MCP Any instance to serve verified skills to any agent (OpenClaw, CrewAI, etc.) regardless of the transport layer.

## 2. Goals & Non-Goals
* **Goals:**
    * Automatically discover and parse `SKILL.md` files in configured directories.
    * Map `SKILL.md` instructions to MCP `prompts` and executable `tools`.
    * Validate skill signatures to ensure they haven't been tampered with (Supply Chain Integrity).
    * Support dynamic reloading of skills when the underlying `.md` file changes.
* **Non-Goals:**
    * Executing arbitrary code embedded in Markdown (tools must still point to verified upstream adapters).
    * Converting `SKILL.md` into other formats like `README.md`.

## 3. Critical User Journey (CUJ)
* **User Persona:** Agent Platform Engineer
* **Primary Goal:** Share a "Security Audit" skill across a swarm of 50 agents running in different environments.
* **The Happy Path (Tasks):**
    1. Engineer places `security-audit.SKILL.md` into the MCP Any skill directory.
    2. MCP Any detects the new file and parses the YAML front-matter and instructions.
    3. The skill is registered as a new MCP tool `call_skill_security_audit`.
    4. An OpenClaw agent calls the tool; MCP Any provides the formatted system prompt and required sub-tools.
    5. The agent executes the task according to the centralized skill definition.

## 4. Design & Architecture
* **System Flow:**
    `SKILL.md File` -> `Skill Parser` -> `MCP Tool Registry` -> `Agent (JSON-RPC)`
    1. **Parsing Engine**: A Go-based Markdown parser that extracts metadata (name, version, inputs) from front-matter and steps from the body.
    2. **Skill-to-Tool Mapping**: Each skill is wrapped in a dynamic tool schema. The tool's `description` is pulled from the skill's summary.
    3. **Runtime Execution**: When called, the adapter hydrates the skill's template with the provided arguments and returns it as a tool result or system prompt.
* **APIs / Interfaces:**
    * New Adapter Type: `skill_adapter`
    * Config: `skills_path: "/path/to/skills"`
* **Data Storage/State:**
    * In-memory cache of parsed skills, keyed by SHA256 hash of the file content.

## 5. Alternatives Considered
* **Manual Conversion to YAML**: Rejected because it breaks compatibility with the existing `SKILL.md` ecosystem used by Claude Code.
* **Agent-side Parsing**: Rejected because it prevents centralized security validation and auditing.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: Skills must be signed. Unsigned skills are "Quarantined" and require explicit approval in the `Project Config Attestation Dashboard`.
* **Observability**: Skill usage (invocation count, success rate) will be tracked and displayed in the `Tool Usage Analytics` dashboard.

## 7. Evolutionary Changelog
* **2026-03-11:** Initial Document Creation.
