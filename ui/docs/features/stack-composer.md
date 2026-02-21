# Intelligent Stack Composer

**Status:** Implemented

## Goal

Transform "Config-as-Code" into a visual composition experience. The Stack Composer allows users to assemble complex microservice architectures using an interactive palette and intelligent YAML editor.

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

- **Auto-Complete**: Press `Ctrl+Space` to see available fields (like `services`, `version`, `image`).
- **Validation**: Basic YAML syntax highlighting is provided. Full schema validation occurs upon saving.

### 4. Deploy

Once satisfied, click **"Save & Deploy"**. The system will validate and save the configuration. If the server is configured for hot-reload, the changes will be applied automatically to the running instance.
