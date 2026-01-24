# Prompts Library

**Status:** Implemented

## Goal
Discover and use pre-defined prompt templates. Servers can expose standardized prompts (e.g., "Analyze Code", "Summarize Text") to be used by clients or the Playground.

## Usage Guide

### 1. Browse Prompts (Workbench)
Navigate to `/prompts`. The interface is divided into two panes:
- **Left Sidebar**: A searchable list of available prompts from all connected services.
- **Right Pane**: The Prompt Workbench for the selected prompt.

![Prompts Workbench](screenshots/prompts_list.png)

### 2. Configure & Generate
Select a prompt from the list to view its details.
- **Configuration**: Enter values for the required arguments in the form.
- **Generate**: Click **"Generate Preview"** to execute the prompt template and produce the messages.

### 3. Output & Usage
The generated messages are displayed in the **Output Preview** section.
- **Copy**: Use the copy button to copy the result to your clipboard.
- **Playground**: Click **"Open in Playground"** to navigate to the Playground, where you can paste the result or start a new session.
