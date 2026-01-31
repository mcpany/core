# Intelligent Stack Composer

**Status:** Alpha / In Development

## Goal

Transform "Config-as-Code" into a visual composition experience. The Stack Composer allows users to assemble complex microservice architectures using a drag-and-drop palette and intelligent YAML editor.

> **Note**: The Stack Composer is currently in an early Preview state. The full visual editor is under active development.

## Usage Guide

### 1. View Stacks

Navigate to `/stacks`. This view lists the currently active configuration stacks (e.g., the system stack).

### 2. Stack Editor (Upcoming)

*Future feature:* The editor will be divided into three panes:
- **Left**: Service Palette (Templates).
- **Center**: YAML Editor.
- **Right**: Live Visualizer.

![Stack Composer Overview](screenshots/stack_composer_overview.png)

### 3. Using the Palette (Upcoming)

To add a service:
1. Open the **Service Palette**.
2. Click on a template (e.g. `Postgres`, `Redis`).
3. The corresponding YAML configuration is injected into the editor cursor position.

### 4. Deploy

Once satisfied, click **"Save Changes"**. The system will save the configuration.
