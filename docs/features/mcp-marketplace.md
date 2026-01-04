# MCP Service Marketplace

The **MCP Service Marketplace** provides a user-friendly, one-click installation experience for Model Context Protocol (MCP) servers. It allows users to browse a catalog of popular, certified MCP servers and install them without needing to manually construct command-line arguments or JSON configurations.

![MCP Marketplace Screenshot](../../.audit/ui/2025-01-04/mcp_marketplace.png)

## Features

-   **Catalog Browsing**: View a curated list of MCP servers categorized by function (Database, Filesystem, Productivity, etc.).
-   **Search & Filter**: Quickly find services by name or description.
-   **One-Click Install**: Install services with a simplified configuration dialog.
-   **Environment Configuration**: The marketplace handles parameter substitution (e.g., `${DB_PATH}`) and prompts for required secrets (e.g., `GITHUB_TOKEN`).

## Architecture

The marketplace is implemented as a frontend component (`ServiceMarketplace`) that consumes a static definition list (`MARKETPLACE_ITEMS`). When an installation is confirmed, it uses the `apiClient.registerService` method to register the new service configuration with the backend.

### Data Structure

Each marketplace item is defined by the `MarketplaceItem` interface:

```typescript
export interface MarketplaceItem {
    id: string;
    name: string;
    description: string;
    config: {
        command: string;
        args: string[]; // Supports ${VAR_NAME} substitution
        envVars: {
            name: string;
            required: boolean;
            // ...
        }[];
    };
    // ...
}
```

## Supported Services

Currently supported services include:

-   **Filesystem**: Secure local directory access.
-   **SQLite**: Local database querying.
-   **GitHub**: Repository and issue management.
-   **PostgreSQL**: Remote database access.
-   **Google Drive**: File access.
-   **Slack**: Messaging integration.
-   **Brave Search**: Web search.
