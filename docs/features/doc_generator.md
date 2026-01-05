# Documentation Generator

**Status**: Implemented

The Documentation Generator automatically produces Markdown documentation for all registered tools and their input schemas. This ensures your documentation never drifts from the actual implementation.

## How it works

The generator:
1.  Loads the server configuration.
2.  Initializes all upstream services (without starting them).
3.  Discovers all tools from each upstream.
4.  Generates a Markdown file listing each tool, its description, and its input parameters (JSON Schema).

## Usage

This feature is typically invoked programmatically via `server/pkg/config/doc_generator.go`.

*(Future versions may expose this via `mcpctl`)*
