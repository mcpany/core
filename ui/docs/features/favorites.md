# Favorites / Pinned Tools

**Date:** 2026-01-18
**Author:** Jules (Lead Engineer)

## Overview

The "Favorites/Pinned Tools" feature allows users to mark specific tools as favorites for quick access. This is particularly useful in environments with many registered tools, where searching or scrolling can be inefficient.

## Key Features

1.  **Pin Toggle:** Users can click a "Pin" icon next to any tool in the list or in the Tool Inspector.
2.  **Pinned Section:** A dedicated "Pinned Tools" section appears at the top of the Tools page, showing all favorited tools.
3.  **Persistence:** Pinned tools are saved in the browser's `localStorage`, ensuring they persist across sessions.
4.  **Visual Feedback:** Pinned tools are highlighted with a yellow star icon.

## Implementation Details

-   **Context:** `FavoritesContext` (`ui/src/contexts/favorites-context.tsx`) manages the state of pinned tools.
-   **UI Components:**
    -   `ToolsPage` (`ui/src/app/tools/page.tsx`): Renders the "Pinned Tools" section and integrates the pin button.
    -   `ToolInspector` (`ui/src/components/tools/tool-inspector.tsx`): Includes a pin button in the dialog header.

## Screenshots

![Favorites Pinned Tools](favorites_pinned_tools.png)
