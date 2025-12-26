# Global Search (Command Palette)

## Overview

The Global Search (Command Palette) is a new feature that brings a "Premium Enterprise" feel to the MCP Any UI. It allows users to quickly navigate between different sections of the application (Dashboard, Playground, Services, Logs, Settings) and perform actions like toggling the theme, all without leaving their keyboard.

## Features

-   **Keyboard Shortcut:** Accessible via `Cmd+K` (macOS) or `Ctrl+K` (Windows/Linux).
-   **Navigation:** Quick jump to:
    -   Dashboard
    -   Playground
    -   Services
    -   Logs
    -   Settings
-   **Theme Switching:** Toggle between Light, Dark, and System themes.
-   **Search:** Filter commands by typing.

## Implementation Details

-   **Library:** `cmdk` (via `shadcn/ui` components).
-   **Components:**
    -   `ui/src/components/ui/command.tsx`: Base command components.
    -   `ui/src/components/command-palette.tsx`: The main palette component implementing the logic.
-   **Integration:** Mounted globally in `ui/src/app/layout.tsx` to ensure availability on all pages.

## Verification

The feature has been verified with:
-   **Unit Tests:** `ui/tests/unit/command-palette.test.tsx` ensuring rendering and logic correctness.
-   **E2E Tests:** `ui/tests/e2e/command-palette.spec.ts` ensuring browser interaction and navigation.

![Global Search Screenshot](.audit/ui/2025-12-26/global_search_snake_case.png)
