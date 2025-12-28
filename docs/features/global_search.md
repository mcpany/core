# Global Search (Command Palette)

## Overview

The Global Search (also known as the Command Palette) provides a quick, keyboard-accessible way to navigate the application, execute commands, and manage settings.

## Features

-   **Keyboard Shortcut:** Press `Cmd+K` (Mac) or `Ctrl+K` (Windows/Linux) to open.
-   **Navigation:** Jump instantly to:
    -   Dashboard
    -   Services
    -   Logs
    -   Playground
    -   Settings
-   **Theme Switching:** Quickly toggle between Light, Dark, and System themes.
-   **Tool Execution:** (Placeholder) Prepared to support executing registered tools directly.

## Implementation Details

-   **Component:** `GlobalSearch` in `ui/src/components/global-search.tsx`.
-   **Library:** Built on top of `cmdk` (Command K), styled with `shadcn/ui` principles and Tailwind CSS.
-   **Integration:** Injected into `ui/src/app/layout.tsx` so it is available on every page.

## Verification

![Global Search Open](global_search_open.png)
