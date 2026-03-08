# Design Doc: Hierarchical Project-Scoped Configuration

**Status:** Draft
**Created:** 2026-03-08

## 1. Context and Scope
Standard MCP adapters rely on a single global configuration file (e.g., `mcpany.yaml`). This creates friction for developers who need project-specific tools, secrets, or environment variables that shouldn't be shared globally or committed to a shared repository. Inspired by `git` and `claude_desktop` patterns, MCP Any needs a hierarchical configuration model that allows local overrides while maintaining global defaults.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Support a `.mcpany/local.yaml` file in the current working directory.
    *   Implement a "Merge Strategy" where local settings override global ones (e.g., environment variables, tool permissions).
    *   Enable "Local-Only" tool registration for project-specific scripts.
    *   Ensure secrets in local configs are never accidentally leaked to global logs.
*   **Non-Goals:**
    *   Implementing a full distributed configuration system (like Consul/Etcd).
    *   Auto-detecting project types (focus only on the presence of the `.mcpany` directory).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Full-stack developer working on multiple client projects.
*   **Primary Goal:** Use project-specific MCP tools (like a local database explorer) without polluting their global tool list or leaking client secrets.
*   **The Happy Path (Tasks):**
    1.  User creates a `.mcpany/config.yaml` in their project root.
    2.  They define a local MCP server (e.g., `sqlite-local`).
    3.  When running MCP Any from that directory, the local tools are merged with global tools.
    4.  Local environment variables (e.g., `DATABASE_URL`) take precedence over global ones.
    5.  The developer commits their code but `.mcpany/local.yaml` is git-ignored.

## 4. Design & Architecture
*   **System Flow:**
    - **Config Discovery**: On startup/reload, MCP Any scans the current directory for `.mcpany/config.yaml`.
    - **Hierarchical Merge Engine**:
        1. Load `~/.mcpany/global.yaml` (Base).
        2. Load `./.mcpany/config.yaml` (Project).
        3. Load `./.mcpany/local.yaml` (Developer specific - Ignored).
    - **Layered Validation**: Each layer is validated independently before merging to ensure structural integrity.
*   **APIs / Interfaces:**
    - New CLI flag: `--project-root [path]` to force a specific scope.
    - Config Metadata: `_mcp_config_source: "global" | "project" | "local"`.
*   **Data Storage/State:** Multi-layered configuration tree in memory.

## 5. Alternatives Considered
*   **Environment Variables Only**: Using `MCP_TOOL_XYZ` variables for overrides. *Rejected* as it becomes unmanageable for complex tool configurations.
*   **Symlinking Configs**: Manually symlinking project configs to a global folder. *Rejected* as it's too high-friction for most developers.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** Local configurations must be subject to the same Zero-Trust Policy Engine as global ones. Project-specific tools shouldn't be able to "elevate" their privileges beyond the global security baseline.
*   **Observability:** The `mcpany doctor` command should list all loaded configuration layers and highlight which values are being overridden.

## 7. Evolutionary Changelog
*   **2026-03-08:** Initial Document Creation.
