# Structured Output Transformation

Transform the output of your tools using powerful query languages like JQ and JSONPath.

## Overview

Sometimes raw API responses are too verbose or complex for an LLM. Transformation allows you to shape the data *before* it returns to the agent.

## Features

- **JQ Support**: Use JQ filters to slice, dice, and reshape JSON data.
- **JSONPath**: Use standard JSONPath expressions for selection.
- **Templates**: Format output using Go templates.

## Example

```yaml
transform:
  type: jq
  filter: ".items[] | {name: .metadata.name, status: .status.phase}"
```
