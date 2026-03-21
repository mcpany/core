# Theme Builder

The Theme Builder allows users to customize the visual appearance of the MCP Any dashboard.

## Overview

The dashboard supports a flexible theming system that currently includes light and dark modes, with the architecture in place to support custom color schemes in the future.

## Features

- **Dark/Light Mode Toggle**: Seamlessly switch between light and dark themes.
- **System Preference Sync**: Automatically matches the user's system theme settings.
- **Persistent Settings**: Theme preference is saved locally.

## Usage

The theme toggle is located in the dashboard header/sidebar. Click the sun/moon icon to switch modes.

## Implementation

The theming engine is built using `next-themes` and React context.
- Component: `ui/src/components/theme-provider.tsx`
- Toggle: `ui/src/components/theme-toggle.tsx`
