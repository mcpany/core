import { render, screen, fireEvent } from '@testing-library/react'
import { describe, it, expect, vi } from 'vitest'
import { CommandMenu } from './command-menu'

// Mock next/navigation
vi.mock('next/navigation', () => ({
  useRouter: () => ({
    push: vi.fn(),
  }),
}))

// Mock next-themes
vi.mock('next-themes', () => ({
  useTheme: () => ({
    setTheme: vi.fn(),
  }),
}))

// Mock Dialog component since it uses portals which might be tricky in JSDOM without setup
// But Radix usually handles it. Let's try testing the interaction.
// Actually, we need to mock ResizeObserver for Radix UI if it's used, but let's see.
// The CommandDialog opens based on state.

describe('CommandMenu', () => {
  it('is initially closed (not visible)', () => {
    render(<CommandMenu />)
    const input = screen.queryByPlaceholderText('Type a command or search...')
    expect(input).not.toBeInTheDocument()
  })

  it('opens when Cmd+K is pressed', async () => {
    render(<CommandMenu />)

    // Simulate Cmd+K
    fireEvent.keyDown(document, { key: 'k', metaKey: true })

    // Check if it's visible.
    // Note: Radix Dialog might render in a portal. testing-library should find it if it's in the document body.
    const input = await screen.findByPlaceholderText('Type a command or search...')
    expect(input).toBeInTheDocument()
  })

    it('opens when Ctrl+K is pressed', async () => {
    render(<CommandMenu />)

    // Simulate Ctrl+K
    fireEvent.keyDown(document, { key: 'k', ctrlKey: true })

    const input = await screen.findByPlaceholderText('Type a command or search...')
    expect(input).toBeInTheDocument()
  })
})
