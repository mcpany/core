# Merge Strategy and Profile-Based Tool Selection

This document describes the new configuration merge strategies and profile-based tool selection capabilities options to give users more control over how configurations are combined and which tools are active.

## Merge Strategy

When loading configurations from multiple sources (e.g., base config and overlay config), you can now control how lists and maps are merged.

### Configuration

The `merge_strategy` field in `McpBundleConfig` controls this behavior.

```yaml
merge_strategy:
  # "extend" (default) or "replace"
  tool_list: extend
  # "extend" (default) or "replace"
  profile_list: extend
  # "extend" (default) or "replace"
  mcp_server_list: extend
```

### Modes

- **`extend` (Default)**:
  - **Lists**: The new list is appended to the existing list.
  - **Maps**: New keys are added. Existing keys are overwritten (deep merge is NOT supported for values yet, they are replaced).
- **`replace`**:
  - **Lists**: The new list completely replaces the existing list.
  - **Maps**: The new map completely replaces the existing map.

### Example

**Base Config:**

```yaml
tools:
  - name: "base-tool"
```

**Overlay Config (Extend):**

```yaml
merge_strategy:
  tool_list: extend
tools:
  - name: "overlay-tool"
```

**Result:** `[base-tool, overlay-tool]`

**Overlay Config (Replace):**

```yaml
merge_strategy:
  tool_list: replace
tools:
  - name: "overlay-tool"
```

**Result:** `[overlay-tool]`

## Profile-Based Tool Selection

Tools can be associated with "profiles" (e.g., `dev`, `prod`, `readonly`). When starting the server, you specify which profiles to enable. Only tools belonging to enabled profiles (or available to all if no profiles are enforcing) are loaded.

### 1. Explicit Association

Tools can explicitly list the profiles they belong to in their definition.

```yaml
tools:
  - name: "dangerous-tool"
    profiles: ["admin", "dev"]
```

If you start the server with `--profiles admin`, this tool is included.

### 2. Dynamic Association (Selectors)

You can define profiles that dynamically select tools based on **Tags** or **Properties**.

**Configuration in `GlobalSettings`:**

```yaml
profile_definitions:
  - name: "safe-mode"
    selector:
      tags: ["safe"]
      tool_properties:
        destructive: "false"
        read_only: "true"
```

**Tool Definition:**

```yaml
tools:
  - name: "read-file"
    tags: ["safe"]
    annotations:
      read_only: true
```

If you start the server with `--profiles safe-mode`, the `read-file` tool is included because it matches the selector criteria.

### Logic Details

A tool is allowed if **ANY** of the following are true:

1. No profiles are enabled on the server (Allow All).
2. The tool explicitly lists one of the enabled profiles.
3. The tool matches the selector of one of the enabled profiles.

Selectors match if:

- **Tags**: The tool has **AT LEAST ONE** of the selector's tags (if selector tags are present).
- **Properties**: The tool matches **ALL** specified properties (e.g. `read_only=true`) (if selector properties are present).
- If both are present, typically both must match.

## CLI Usage

```bash
# Enable specific profiles
gemini start --profiles dev,safe-mode
```
