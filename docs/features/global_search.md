# Global Search (Command Palette)

## Overview
A "Spotlight-style" command palette accessible via `Cmd+K` (or `Ctrl+K`) that provides instant access to all resources in the MCP Any platform.

## Features
- **Global Shortcut:** Access from anywhere with `Cmd+K`.
- **Sidebar Access:** A prominent search button in the sidebar.
- **Navigation:** Jump to Dashboard, Services, Tools, Logs, Settings, etc.
- **Visuals:** Blurred backdrop, glassmorphism, consistent with the "Unifi/Apple" design language.

## Screenshot
![Global Search](global_search.png)

## Technical Details
- Built with `cmdk` and `radix-ui` primitives.
- Integrated into the global `layout.tsx`.
- Refactored `Sidebar` to be globally persistent.
