# Prompts Library & Workbench

**Status:** Implemented

## Goal
Discover and use pre-defined prompt templates. Servers can expose standardized prompts (e.g., "Analyze Code", "Summarize Text") to be used by clients or the Playground.

## Usage Guide

### 1. Browse Prompts
Navigate to `/prompts`.
The **Prompt Library** (left pane) lists all available prompts exposed by upstream services.
- Use the search bar to filter prompts by name or description.
- Click on a prompt to select it.

### 2. Configure & Preview
Once a prompt is selected, the **Workbench** (right pane) displays its details.
- **Configuration**: If the prompt requires arguments, a form is generated based on its input schema. Fill in the required fields.
- **Generate Preview**: Click the "Generate Preview" button to execute the prompt against the server.
- **Output Preview**: The resulting messages (User/Assistant) are displayed in the preview area.

### 3. Use in Playground
After generating a preview, you can transfer the result to the main Playground.
- Click **"Open in Playground"** (appears after successful generation).
- This will redirect you to `/playground` with the prompt result pre-loaded for further interaction.
