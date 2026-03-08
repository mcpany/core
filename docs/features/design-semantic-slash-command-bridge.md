# Design Doc: Semantic Slash-Command Bridge

**Status:** Draft
**Created:** 2026-03-08

## 1. Context and Scope
Agent-to-Tool interaction is evolving from raw JSON-RPC calls to semantic commands. Platforms like Gemini CLI and Claude Code are increasingly using "slash commands" (e.g., `/search`, `/edit`) to trigger complex agent behaviors. Currently, MCP Any exposes MCP Prompts as static resources. To maintain parity with native CLI experiences, MCP Any must bridge the gap between MCP Prompts and CLI-native slash commands, allowing agents to "discover" and "execute" these prompts using their preferred semantic interface.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Automatically map MCP Prompts to structured Slash Commands.
    *   Generate metadata for slash commands (description, arguments) from Prompt schemas.
    *   Provide a standard "Slash-Command Registry" for CLI agents.
    *   Support dynamic argument binding from CLI inputs to Prompt parameters.
*   **Non-Goals:**
    *   Implementing a new CLI shell (this is a bridge for *existing* CLI agents).
    *   Replacing the standard MCP Prompt protocol.

## 3. Critical User Journey (CUJ)
*   **User Persona:** CLI Agent Developer (e.g., contributing to Gemini CLI extensions).
*   **Primary Goal:** Use an MCP Prompt named `explain_code` as a native `/explain` command in the terminal.
*   **The Happy Path (Tasks):**
    1.  User connects an MCP server to MCP Any that exposes an `explain_code` prompt.
    2.  MCP Any detects the prompt and registers it in the Semantic Bridge.
    3.  The CLI Agent queries MCP Any for available commands.
    4.  The CLI Agent displays `/explain [code]` to the user.
    5.  User types `/explain context: "func main() { ... }"`
    6.  The Semantic Bridge translates this into an MCP `prompts/get` call with the provided arguments.

## 4. Design & Architecture
*   **System Flow:**
    - **Introspection**: The Bridge crawls all connected MCP servers for `prompts/list`.
    - **Normalization**: Prompts are converted into a "Semantic Command Model" (SCM).
    - **Dynamic Proxy**: When a slash command is "executed," the Bridge proxies the call to the underlying MCP `prompts/get` implementation.
*   **APIs / Interfaces:**
    - `/api/v1/commands`: Returns a list of all mapped slash commands.
    - `/api/v1/commands/execute`: Endpoint for CLI agents to trigger the command.
*   **Data Storage/State:** In-memory registry of mappings, refreshed on MCP server reload.

## 5. Alternatives Considered
*   **Manual Mapping**: Requiring users to manually define slash commands in `config.yaml`. *Rejected* as it adds too much friction compared to Gemini CLI's Auto-Discovery.
*   **Hardcoded Aliases**: *Rejected* in favor of a dynamic schema-based approach.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** Commands are subject to the same Policy Firewall as tool calls. Access to specific prompts can be restricted via capability tokens.
*   **Observability:** Logs will reflect whether a prompt was accessed via the standard MCP protocol or the Semantic Bridge.

## 7. Evolutionary Changelog
*   **2026-03-08:** Initial Document Creation.
