# Global Search (Command Palette)

## Overview
A new "Apple-style" Command Palette (`Cmd+K`) has been implemented to improve navigation and accessibility across the MCP Any dashboard.

## Features
- **Keyboard Shortcut**: Accessible globally via `Cmd+K` (or `Ctrl+K`).
- **Sidebar Trigger**: A visual search button in the sidebar.
- **Navigation**: Quickly jump to:
    - Dashboard
    - Services
    - Tools
    - Logs
    - Settings
    - Playground
- **Settings**: Quick access to API Keys and General Settings.
- **Theming**: Placeholder for future theme switching.
- **Help**: Links to GitHub and Documentation.

## Implementation Details
- **Component**: `ui/src/components/command-menu.tsx`
- **Library**: `cmdk` (wrapped in Shadcn/UI style components).
- **Styling**: Uses Tailwind CSS with blurred backdrop and consistent typography.

## Verification
- **Unit/E2E Tests**: `ui/tests/command-palette.spec.ts` covers opening, filtering, and navigation.
- **Screenshot**:
![Command Palette](.audit/ui/2025-05-18/global_search_command_palette.png)
