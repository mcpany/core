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
