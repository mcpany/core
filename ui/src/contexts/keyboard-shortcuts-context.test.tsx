/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React from "react"
import { render, screen, fireEvent } from "@testing-library/react"
import { vi, describe, it, expect } from "vitest"
import { KeyboardShortcutsProvider, useShortcut, useKeyboardShortcuts } from "./keyboard-shortcuts-context"

// Mock component to test the hook
const TestComponent = ({ action }: { action: () => void }) => {
  useShortcut("test.shortcut", ["meta+k"], action, { label: "Test Shortcut" })
  return <div>Test Component</div>
}

const TestOverrideComponent = () => {
  const { updateOverride, getKeys } = useKeyboardShortcuts()
  return (
    <div>
      <div data-testid="keys">{getKeys("test.shortcut").join(",")}</div>
      <button onClick={() => updateOverride("test.shortcut", ["shift+x"])}>Override</button>
    </div>
  )
}

describe("KeyboardShortcutsContext", () => {
  it("should trigger action on default key press", () => {
    const action = vi.fn()
    render(
      <KeyboardShortcutsProvider>
        <TestComponent action={action} />
      </KeyboardShortcutsProvider>
    )

    fireEvent.keyDown(window, { key: "k", metaKey: true })
    expect(action).toHaveBeenCalledTimes(1)
  })

  it("should not trigger action on wrong key press", () => {
    const action = vi.fn()
    render(
      <KeyboardShortcutsProvider>
        <TestComponent action={action} />
      </KeyboardShortcutsProvider>
    )

    fireEvent.keyDown(window, { key: "b", metaKey: true })
    expect(action).not.toHaveBeenCalled()
  })

  it("should handle overrides", () => {
    const action = vi.fn()
    render(
      <KeyboardShortcutsProvider>
        <TestComponent action={action} />
        <TestOverrideComponent />
      </KeyboardShortcutsProvider>
    )

    // Initial check
    expect(screen.getByTestId("keys")).toHaveTextContent("meta+k")

    // Apply override
    fireEvent.click(screen.getByText("Override"))
    expect(screen.getByTestId("keys")).toHaveTextContent("shift+x")

    // Default should not work anymore
    fireEvent.keyDown(window, { key: "k", metaKey: true })
    expect(action).not.toHaveBeenCalled()

    // Override should work
    fireEvent.keyDown(window, { key: "X", shiftKey: true }) // shift sends capital X usually, or matchesKey logic handles it
    expect(action).toHaveBeenCalledTimes(1)
  })
})
