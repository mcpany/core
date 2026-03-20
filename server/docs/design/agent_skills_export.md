# Design: Agent Skills Export & Role-Based Access

## Problem Statement
We have a single `mcpany` server instance supporting 1000 users. We need to export different sets of "Agent Skills" (MCP Tools) to different user roles.
- **Scale**: 1000 users, 1 server.
- **Target**: Enterprise users (and regular users).
- **Core Requirement**: "Export different set of agents skills to different role of users".
- **Proposed Idea**: "Directory remote mount" (or similar file-based abstraction).

## Goals
1.  **Isolation**: User A (Role A) should only see/execute Tool Set A. User B (Role B) sees Tool Set B.
2.  **Compatibility**: Support standard MCP clients (Claude Desktop, etc.) which often rely on `stdio` or simple config.
3.  **Ease of Deployment**: Avoid 1000 separate containers. Single multi-tenant server.

## Proposed Solution: Role-Based Virtual Mounts

### 1. Conceptual Architecture
The `mcpany` server acts as a "Virtual Tool Filesystem".
Instead of just serving MCP over HTTP/WS, we introduce a **Mountable Interface** (or a client-side bridge) that projects the allowed tools for a user into a local directory structure.

### 2. User Experience
1.  **SysAdmin/DevOps**: Defines Roles in `mcpany` config.
    ```yaml
    roles:
      - name: "developer"
        tools: ["git-commit", "postgres-query"]
      - name: "analyst"
        tools: ["sql-query", "splunk-search"]
    ```
2.  **User**:
    - **Option A (Network Mount)**: Mounts `mcpany:/export/developer` to `/mnt/mcp-skills`.
    - **Option B (Client Bridge)**: Runs `mcpany connect --role developer --mount ./my-skills`.
3.  **Client (e.g., Claude Desktop)**:
    - Configured to look at `/mnt/mcp-skills`.
    - Sees executable stubs: `git-commit`, `postgres-query`.
    - Executes them. The stubs proxy the request to `mcpany` (via locally bridged socket or HTTPS).

### 3. Implementation Details

#### Role-Based Access Control (RBAC)
- Reuse existing "Service Profiles" or extend them.
- **Config**: Map `User Identity` (from OIDC/Header) -> `Profile` -> `Allowed Tools`.
- **Enforcement**: Middleware in `mcpany` checks `x-mcp-role` or JWT claims.

#### Export Mechanism: The "Directory Mount"
To support "Exporting" to a directory, we can use FUSE or a Sync Client.
Given "Directory remote mount" suggestion:
- **SSHFS/NFS**: Doable but complex to secure per-user on 1 port.
- **WebDAV**: Common, mountable on OS.
- **FUSE (Filesystem in Userspace)**: `mcpany-fuse` binary.
    - `mcpany-fuse mount https://mcpany-server/api/v1 --token <user-token> /local/mountpoint`
    - It lists tools as "Files" (Executables).
    - Executing the file triggers the MCP RPC.

#### Alternative: "Virtual" Agent Export
If the client is MCP-native (like Claude Desktop), we don't strictly *need* a directory, but Claude Desktop config (`claude_desktop_config.json`) expects *local commands* for `stdio`.
- **The "Stub Generator"**:
    - `mcpany export-skills --role developer --out ./skills-dir`
    - Creates shell scripts in `./skills-dir` for each allowed tool.
    - Script content:
        ```bash
        #!/bin/bash
        # Wrapper for 'git-commit' tool
        mcpany invoke --server https://mcpany-server --tool git-commit "$@"
        ```
    - User points Claude Desktop to these scripts.

### 4. Enterprise Readiness
- **Audit Logs**: Track which User/Role invoked which tool.
- **Dynamic Updates**: If admin adds a tool to "developer" role, `export-skills` (or FUSE mount) updates automatically.

## Recommendation
1.  **Implement RBAC** in `mcpany` (extend Service Profiles).
2.  **Build `mcpany export-skills` CLI**: A simple "stub generator" that fetches allowed tools and creates wrapper scripts. This typically satisfies "Directory based" usage for `stdio` clients without complex FUSE mounting.
3.  **Investigate FUSE**: For advanced "Live" mounting if static stubs are insufficient.
