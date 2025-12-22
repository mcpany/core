import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { CommandMenu } from './command-menu'
import { vi, describe, it, expect } from 'vitest'

// Mock useRouter
const mockPush = vi.fn()
vi.mock('next/navigation', () => ({
  useRouter: () => ({
    push: mockPush,
  }),
}))

describe('CommandMenu', () => {
  it('opens on Cmd+K', async () => {
    render(<CommandMenu />)

    // Dialog should be closed initially
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()

    // Press Cmd+K
    fireEvent.keyDown(document, { key: 'k', metaKey: true })

    // Check if dialog is open
    await waitFor(() => expect(screen.getByRole('dialog')).toBeInTheDocument())
  })

  it('navigates when item is selected', async () => {
    render(<CommandMenu />)

    // Open it
    fireEvent.keyDown(document, { key: 'k', metaKey: true })

    // Wait for dialog
    await waitFor(() => expect(screen.getByRole('dialog')).toBeInTheDocument())

    // Type to filter
    const input = screen.getByPlaceholderText('Type a command or search...')
    fireEvent.change(input, { target: { value: 'stripe' } })

    // Click stripe_charge
    const stripeItem = await screen.findByText('stripe_charge')
    fireEvent.click(stripeItem)

    // Check navigation
    expect(mockPush).toHaveBeenCalledWith('/tools')
  })
})
