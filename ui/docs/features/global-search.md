# Global Search Feature

## Overview

The Global Search feature provides a centralized command palette for quick navigation and action execution within the MCP Any UI. Inspired by modern "Spotlight" or "Cmd+K" interfaces, it allows users to jump to services, tools, resources, prompts, and settings without leaving the keyboard.

## Features

- **Keyboard Shortcut:** Accessible via `Cmd+K` (macOS) or `Ctrl+K` (Windows/Linux).
- **Universal Search:** Search across:
  - **Navigation:** Dashboard, Services, Tools, Resources, Logs, Playground, Settings.
  - **Services:** Jump directly to a specific service detail page.
  - **Tools:** Find and navigate to tool details.
  - **Resources:** Access resource definitions.
  - **Prompts:** Quickly find available prompts.
- **Theme Switching:** Toggle between Light, Dark, and System themes directly from the palette.
- **Glassmorphism Design:** Matches the "Unifi" / "Apple" aesthetic with blurred backdrops and subtle borders.

## Implementation Details

The feature is implemented using the `cmdk` library for the accessible command menu primitive and `radix-ui` for the dialog structure. It leverages `next/navigation` for client-side routing and `next-themes` for theme management.

### Key Components

- `GlobalSearch` (`ui/src/components/global-search.tsx`): The main component containing the dialog logic and data fetching.
- `apiClient` (`ui/src/lib/client.ts`): Used to fetch the list of services, tools, resources, and prompts dynamically when the palette is opened.

## Usage

1. Press `Cmd+K` or click the "Search..." button in the top navigation bar.
2. Type to filter results.
3. Use `Up`/`Down` arrows to navigate.
4. Press `Enter` to select an item.

## Screenshot

![Global Search](.audit/ui/2025-12-31/global_search.png)
