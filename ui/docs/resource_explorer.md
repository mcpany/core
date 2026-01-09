# Resource Explorer

**Date:** 2026-01-09
**Feature:** Resource Explorer Component

## Overview
The Resource Explorer transforms the previous basic list of resources into a full-featured, Finder-style management interface. This enhancement aligns with the goal of creating a "Premium Enterprise" console for MCP Any.

## Key Features
- **Split-View Interface:** A master-detail layout allowing users to browse resources on the left and preview them on the right.
- **Rich Previews:**
  - Syntax highlighting for JSON, YAML, XML, and Code files.
  - Markdown rendering for documentation.
- **Search & Filter:** Real-time filtering of resources by name or URI.
- **View Modes:** Toggle between List and Grid views for better usability.
- **Actions:** Quick access to "Copy Content" and "Download" for each resource.

## Screenshot
![Resource Explorer](../../.audit/ui/2026-01-09/resource_explorer.png)

## Implementation Details
- **Component:** `ui/src/components/resources/resource-explorer.tsx`
- **Route:** `ui/src/app/resources/page.tsx`
- **Client Library:** Updated `ui/src/lib/client.ts` with `readResource` support (mocked for UI demo).
