# Global Search (Cmd+K)

The Global Search feature provides a centralized command palette for quick navigation, resource access, and actions across the MCP Any application.

## Features

- **Quick Navigation:** Instantly jump to any page (Dashboard, Services, Resources, Tools, Prompts, Settings).
- **Service Search:** Search through all registered upstream services and navigate directly to their details.
- **Keyboard Shortcut:** Accessible from anywhere using `Cmd+K` (Mac) or `Ctrl+K` (Windows/Linux).
- **Theme Support:**  (Planned) Quick switching between Light, Dark, and System themes.

## Screenshots

![Global Search Open](global_search_open.png)

*The Global Search dialog opened with Cmd+K*

![Global Search Filtered](global_search_filtered.png)

*Filtering results by typing "Services"*

## Architecture

The feature is built using:
- **`cmdk`**: A composable, unstyled command menu for React.
- **`shadcn/ui`**: For styling the dialog and input components to match the "Unifi/Apple" design system.
- **`GlobalSearch` Component**: A self-contained component integrated into the `RootLayout` ensuring global availability.
- **`apiClient`**: Dynamically fetches the list of services when the menu is opened.

## Usage

1. Press `Cmd+K` or `Ctrl+K`.
2. Type to filter results.
3. Use `Arrow Up/Down` to navigate.
4. Press `Enter` to select an item.
