#!/usr/bin/env python3
"""
Apply repository-specific compatibility tweaks to generated ts-proto files.

These edits are intentionally deterministic so `make gen` can be rerun after
proto/protobuf updates without requiring manual patching of generated files.
"""

from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]


def patch_file(rel_path: str, replacements: list[tuple[str, str]]) -> None:
    path = ROOT / rel_path
    if not path.exists():
        return
    text = path.read_text()
    original = text
    if "// @ts-nocheck\n/* eslint-disable */" not in text:
        text = text.replace("/* eslint-disable */", "// @ts-nocheck\n/* eslint-disable */", 1)
    for old, new in replacements:
        text = text.replace(old, new)
    if text != original:
        path.write_text(text)


patch_file(
    "proto/config/v1/auth.ts",
    [
        ("  validationRegex: string;\n", "  validationRegex?: string | undefined;\n  /** UI-only reference for selecting a stored secret. */\n  secretId?: string | undefined;\n"),
    ],
)

patch_file(
    "proto/config/v1/call.ts",
    [
        ("  extractionRules: { [key: string]: string };\n", "  extractionRules?: { [key: string]: string } | undefined;\n"),
        ("  template: string;\n", "  template?: string | undefined;\n"),
        ("  jqQuery: string;\n", "  jqQuery?: string | undefined;\n"),
    ],
)

patch_file(
    "proto/config/v1/resource.ts",
    [
        ("  title: string;\n", "  title?: string | undefined;\n"),
        ("  description: string;\n", "  description?: string | undefined;\n"),
        ("  size: Long;\n", "  size?: Long | undefined;\n"),
        ("  disable: boolean;\n", "  disable?: boolean | undefined;\n"),
        ("  profiles: Profile[];\n", "  profiles?: Profile[] | undefined;\n"),
    ],
)

patch_file(
    "proto/config/v1/skill.ts",
    [
        ("  license: string;\n", "  license?: string | undefined;\n"),
        ("  metadata: { [key: string]: string };\n", "  metadata?: { [key: string]: string } | undefined;\n"),
    ],
)

patch_file(
    "proto/config/v1/tool.ts",
    [
        ("  isStream: boolean;\n", "  isStream?: boolean | undefined;\n"),
        ("  title: string;\n", "  title?: string | undefined;\n"),
        ("  readOnlyHint: boolean;\n", "  readOnlyHint?: boolean | undefined;\n"),
        ("  destructiveHint: boolean;\n", "  destructiveHint?: boolean | undefined;\n"),
        ("  idempotentHint: boolean;\n", "  idempotentHint?: boolean | undefined;\n"),
        ("  openWorldHint: boolean;\n", "  openWorldHint?: boolean | undefined;\n"),
        ("  callId: string;\n", "  callId?: string | undefined;\n"),
        ("  disable: boolean;\n", "  disable?: boolean | undefined;\n"),
        ("  profiles: Profile[];\n", "  profiles?: Profile[] | undefined;\n"),
        ("  mergeStrategy: ToolDefinition_MergeStrategy;\n", "  mergeStrategy?: ToolDefinition_MergeStrategy | undefined;\n"),
        ("  tags: string[];\n", "  tags?: string[] | undefined;\n"),
    ],
)

patch_file(
    "proto/config/v1/upstream_service.ts",
    [
        ("  readOnly: boolean;\n", "  readOnly?: boolean | undefined;\n"),
        ("  configurationSchema: string;\n", "  configurationSchema?: string | undefined;\n"),
        ("  tools: ToolDefinition[];\n", "  tools?: ToolDefinition[] | undefined;\n"),
        ("  resources: ResourceDefinition[];\n", "  resources?: ResourceDefinition[] | undefined;\n"),
        ("  calls: { [key: string]: OpenAPICallDefinition };\n", "  calls?: { [key: string]: OpenAPICallDefinition } | undefined;\n"),
        ("  prompts: PromptDefinition[];\n", "  prompts?: PromptDefinition[] | undefined;\n"),
    ],
)
