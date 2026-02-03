# Resource Explorer

**Status:** Implemented

## Goal
Browse and read static assets exposed by connected MCP servers. Resources can include files, database rows, or any other content addressable by a URI.

## Usage Guide

### 1. Resource List
Navigate to `/resources`.
The list displays all available resources aggregated from all healthy services.
- **URI**: Unique Resource Identifier.
- **MIME Type**: Content type (e.g., `text/plain`, `application/json`).

![Resource List](screenshots/resources_list.png)

### 2. Preview Resource
Click on a resource row (e.g., `file:///etc/hosts`).
- **Text/Code**: Opened in a read-only editor with syntax highlighting.
- **Images**: Displayed in a preview modal.

![Resource Preview](screenshots/resource_preview.png)
