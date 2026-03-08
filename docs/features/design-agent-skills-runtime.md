# Design Doc: Agent Skills Secure Runtime
**Status:** Draft
**Created:** 2026-03-08

## 1. Context and Scope
With the rapid growth of the "Agent Skills" ecosystem (OpenClaw, ClawHub), agents are increasingly using dynamic, portable capabilities. However, the "ClawHavoc" campaign discovered over 300 malicious skills, including crypto drainers. MCP Any must provide a secure way to leverage these skills while protecting the host environment. This feature will bridge the gap between the "Agent Skill" folder standard and the "Model Context Protocol" (MCP) while enforcing strict isolation and provenance checks.

## 2. Goals & Non-Goals
* **Goals:**
    * Natively ingest "Agent Skill" folders (containing `SKILL.md`, `scripts/`, `references/`).
    * Expose skill capabilities as standard MCP tools.
    * Execute skill scripts (Python/Bash/JS) in an isolated container (Docker or WebAssembly).
    * Enforce "Progressive Disclosure" as a security middleware.
    * Perform cryptographic provenance checks for all loaded skills.
* **Non-Goals:**
    * Modifying existing Agent Skill formats or specifications.
    * Becoming a primary "Agent Skill" marketplace (MCP Any remains a gateway).
    * Supporting every possible obscure scripting language (prioritizing Python/JS/Bash).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Orchestrator (e.g., OpenClaw user)
* **Primary Goal:** Use a community-contributed "Meeting Mastery" skill from ClawHub safely.
* **The Happy Path (Tasks):**
    1. User points MCP Any to the "Meeting Mastery" skill folder.
    2. MCP Any reads `SKILL.md` and verifies the cryptographic signature (Provenance Check).
    3. MCP Any registers the skill's capabilities as MCP tools, initially exposing only names and basic descriptions (Progressive Disclosure - Level 1).
    4. LLM requests to use the skill. MCP Any prompts for HITL approval or verifies policy (Progressive Disclosure - Level 2).
    5. MCP Any executes the requested script within a Docker-bound container, isolated from the host (Isolated Execution).
    6. Result is returned to the LLM via the standard MCP response format.

## 4. Design & Architecture
* **System Flow:**
    `LLM -> MCP Any (Gateway) -> Skill Loader -> Provenance Checker -> Policy Engine -> Isolated Runtime (Docker/Wasm) -> MCP Any (Gateway) -> LLM`
* **APIs / Interfaces:**
    * `/v1/skills/register`: Accepts a local path or URL to an Agent Skill folder.
    * `SkillToMCPAdapter`: A new internal adapter that maps `SKILL.md` YAML frontmatter and instruction blocks to MCP `Tool` objects.
* **Data Storage/State:**
    * `SkillRegistry`: In-memory index of active skills, backed by SQLite for persistence.
    * `SkillSessionState`: Temporary state stored in the "Shared KV Store" (Blackboard) for multi-step skills.

## 5. Alternatives Considered
* **Native Execution (No Sandbox):** Rejected due to the high risk of malicious code in unverified marketplace skills (ClawHavoc).
* **Virtual Machines:** Rejected due to high overhead and slow startup for frequent tool calls. Docker/Wasm provides a better balance.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All scripts are restricted to a specific "Skill Sandbox" with no host filesystem access unless explicitly granted via Policy.
* **Observability:** Detailed logs of skill script execution (stdout/stderr) are captured and mapped to MCP Any's diagnostic logs.

## 7. Evolutionary Changelog
* **2026-03-08:** Initial Document Creation.
