# Intelligent Stack Composer

**Status:** UI Prototype (Simulated Persistence)

## Goal

Transform "Config-as-Code" into a visual composition experience. The Stack Composer allows users to assemble complex microservice architectures using a drag-and-drop palette and intelligent YAML editor.

## Usage Guide

### 1. Stack Editor

Navigate to `/stacks`. The editor is divided into three panes:

- **Left**: Service Palette (Templates).
- **Center**: YAML Editor.
- **Right**: Live Visualizer.

![Stack Composer Overview](screenshots/stack_composer_overview.png)

### 2. Using the Palette

To add a service:

1. Open the **Service Palette** (if collapsed).
2. Click on a template (e.g. `Postgres`, `Redis`).
3. The corresponding YAML configuration is injected into the editor cursor position.

![Service Palette](screenshots/stack_composer_palette.png)

### 3. Manual Configuration

You can fine-tune the configuration in the Monaco Editor.

- **Validation**: Detailed error markers appear if you violate the schema.
- **Auto-Complete**: Press `Ctrl+Space` to see available fields.

### 4. Deploy

Once satisfied, click **"Deploy Stack"**. The system will provision the defined services and route them appropriately.

*Note: In the current prototype, persistence is simulated and changes are not saved to the backend.*
