# Global Search (Command Palette)

## Overview

The Global Search (Command Palette) is a new feature that provides quick navigation and action execution across the MCP Any UI. Inspired by modern "Power User" tools like VS Code and Linear, it allows users to access any resource or setting via keyboard shortcuts.

![Global Search](global_search_command_palette.png)

## Features

- **Quick Navigation**: Instantly jump to Dashboard, Services, Tools, Logs, Resources, Stacks, Webhooks, and Settings.
- **Theme Toggling**: Switch between Light, Dark, and System themes directly from the palette.
- **Keyboard Access**: Activated globally via `Cmd+K` (macOS) or `Ctrl+K` (Windows/Linux).
- **Search Filtering**: Filter commands by typing.

## Usage

1. Press `Cmd+K` (or `Ctrl+K`).
2. Type to filter options.
3. Use Arrow keys to navigate.
4. Press `Enter` to select.
5. Press `Escape` to close.

## Implementation Details

- **Component**: `CommandMenu` (`ui/src/components/command-menu.tsx`)
- **Library**: `cmdk` (Radix UI primitive)
- **Styling**: Tailwind CSS (Shadcn UI style)
- **Integration**: Added to `RootLayout` in `ui/src/app/layout.tsx`.
