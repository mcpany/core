# Middleware Pipeline

**Status:** Implemented

## Goal
Configure the interceptors and processing steps that occur between the client and upstream servers. Middleware components handle cross-cutting concerns like Authentication, Rate Limiting, and Logging.

## Usage Guide

### 1. Visual Pipeline
Navigate to `/middleware`. The interface visualizes the request flow from left to right.

![Middleware Pipeline](screenshots/middleware_pipeline.png)

### 2. Configure a Component
Click on any middleware block (e.g., **"Rate Limit"**) to inspect its settings.
- **Enable/Disable**: Toggle the component without removing it.
- **Configuration**: Edit parameters (e.g., "Max Requests per Minute").

### 3. Reorder Pipeline
(If supported) Drag and drop components to change their execution order. For example, moving **Logging** before **Auth** will log even rejected unauthenticated requests.
