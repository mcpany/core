# Skill Manager

The Skill Manager allows defining and managing "Skills" - reusable units of knowledge or capability that can be loaded by the MCP server.

## Overview

Skills are stored as directories containing a `SKILL.md` file. Each skill defines:
- A name (based on the directory name)
- Frontmatter metadata (inputs, outputs, description)
- Instructions (markdown body)
- Optional assets (scripts, references)

## SKILL.md Format

The `SKILL.md` file uses YAML frontmatter followed by the skill content/instructions.

```markdown
---
description: "A skill that does something useful"
inputs:
  - name: "param1"
    description: "The first parameter"
    required: true
---

# Instructions

Your markdown instructions go here.
```

## Directory Structure

Skills are stored in a root directory (configurable).

```text
skills/
  ├── my-skill/
  │   ├── SKILL.md
  │   ├── assets/
  │   └── scripts/
  └── another-skill/
      └── SKILL.md
```

## Constraints

- **Name**: Lowercase alphanumeric and hyphens only. No start/end hyphen. No consecutive hyphens. (Regex: `^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)
- **Paths**: Asset paths must be relative to the skill directory and cannot contain parents (`..`).
