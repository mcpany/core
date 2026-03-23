import json

docs_to_sample = [
    # UI docs
    "ui/docs/features/hitl.md",
    "ui/docs/features/recursive_context.md",
    "ui/docs/features/universal_agent_bus.md",

    # Server docs
    "server/docs/features/hitl.md",
    "server/docs/features/recursive_context.md",
    "server/docs/features/shared_kv_store.md",
    "server/docs/features/granular_scopes.md",
    "server/docs/features/lazy-mcp.md",
    "server/docs/features/dlp.md",
    "server/docs/features/audit_logging.md"
]

print(json.dumps(docs_to_sample, indent=2))
