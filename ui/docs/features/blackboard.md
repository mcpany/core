# Blackboard Isolation Inspector

The Blackboard Isolation Inspector visualizes and debugs Agent-Bound Blackboard data across different "Intent Scopes."

## Overview

The embedded SQLite "Blackboard" acts as a shared Key-Value store, providing reliable state management for multi-agent systems. The Isolation Inspector UI allows administrators to review what keys and values are currently stored in the Blackboard and which agents/intents they belong to.

## Usage

1. Navigate to the **Blackboard** page.
2. View the grid of shared state fragments.
3. Review the `Agent ID`, `Value`, and inferred `Intent` for each entry to ensure there is no state leakage across boundaries.
