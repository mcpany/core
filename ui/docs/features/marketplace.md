# Marketplace

**Status:** Implemented

## Goal
Discover and install pre-configured services and stacks. The Marketplace provides a catalog of "Service Templates" (like Postgres, Redis, Starter Kits) that can be deployed with one click.

## Usage Guide

### 1. Browse Catalog
Navigate to `/marketplace`. The grid displays available templates with their popularity and descriptions.

![Marketplace Grid](screenshots/marketplace_grid.png)

### 2. Install Item
1. Click the **"Install"** or **"Configure"** button on a card (e.g., "Postgres").
2. A configuration modal appears, pre-filled with sensible defaults (Port, Password, etc.).
3. Click **"Deploy"** to provision the service.

![Install Modal](screenshots/marketplace_install_modal.png)

### 3. Share Service Collection (Export)
You can export your current service configurations to share with others.

1. Click the **"Share Your Config"** button (Share icon) in the header.
2. Select the services you want to include.
3. **Secret Handling**: Choose how to handle sensitive environment variables (like API Keys):
   - **Redact Secrets** (Default): Replaces values with `<REDACTED>`. Safe for sharing publicly.
   - **Template Variables**: Replaces values with `${VAR_NAME}`. Useful for creating templates.
   - **Unsafe Export**: Keeps original values. **Warning:** Do not share this file publicly as it contains your credentials.
4. Click **"Generate Configuration"** to generate the YAML.
5. Copy the YAML to your clipboard.

![Safe Share Dialog](screenshots/safe_share_dialog.png)
