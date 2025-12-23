# Global Search Feature

## Overview

The Global Search feature provides a quick and accessible way for users to navigate the application and find resources. Inspired by modern command palettes (like Spotlight or Raycast), it allows users to jump to services, tools, settings, and documentation using a keyboard shortcut or a visual trigger.

## Usage

*   **Keyboard Shortcut:** Press `Cmd + K` (on Mac) or `Ctrl + K` (on Windows/Linux) to open the search dialog.
*   **Visual Trigger:**
    *   **Desktop:** Click the "Search... (Cmd+K)" button in the top right corner.
    *   **Mobile:** Tap the floating search icon in the bottom right corner.

## Implementation Details

*   **Component:** `GlobalSearch` (`ui/src/components/global-search.tsx`)
*   **Library:** `cmdk` (Radix UI primitive for command menus)
*   **Design:**
    *   Uses Shadcn UI's `Command` components.
    *   Glassmorphism effect with backdrop blur.
    *   Responsive design with mobile optimizations.

## Screenshot

![Global Search UI](../../.audit/ui/2025-12-23/global_search_snake_case.png)
