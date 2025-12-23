# Global Search Feature

## Overview
The Global Search feature provides a "Cmd+K" (or "Ctrl+K") command palette for the MCP Any UI, allowing users to quickly navigate to pages, services, and tools.

## Implementation Details
- **Component:** `ui/src/components/global-search.tsx`
- **Dependencies:** `cmdk` (Radix UI compatible command menu)
- **Integration:** Added to `ui/src/app/layout.tsx` to be available globally.

## Features
- **Keyboard Shortcut:** Open with `Cmd+K` or `Ctrl+K`.
- **Navigation:**
    - **Dashboard**: Quick link to the main dashboard.
    - **Settings**: Quick link to the settings page.
    - **Services**: Dynamically fetched list of registered services from `apiClient`.
- **Visuals:** Uses glassmorphism and blurred backdrops consistent with the "Premium Enterprise" design system.

## Screenshot
![Global Search](global_search.png)
