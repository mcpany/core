
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { CommandPalette } from '../../src/components/command-palette'
// Mocking Next.js router
jest.mock('next/navigation', () => ({
  useRouter: () => ({
    push: jest.fn(),
  }),
}))

// Mocking useTheme
jest.mock('next-themes', () => ({
    useTheme: () => ({
        setTheme: jest.fn(),
        theme: 'light'
    }),
    ThemeProvider: ({children}: {children: React.ReactNode}) => <div>{children}</div>
}))


describe('CommandPalette', () => {
  it('does not render initially', () => {
    render(<CommandPalette />)
    expect(screen.queryByPlaceholderText('Type a command or search...')).not.toBeInTheDocument()
  })

  it('opens when Cmd+K is pressed', async () => {
    render(<CommandPalette />)

    fireEvent.keyDown(document, { key: 'k', metaKey: true })

    await waitFor(() => {
        expect(screen.getByPlaceholderText('Type a command or search...')).toBeInTheDocument()
    })
  })

  it('opens when Ctrl+K is pressed', async () => {
    render(<CommandPalette />)

    fireEvent.keyDown(document, { key: 'k', ctrlKey: true })

    await waitFor(() => {
        expect(screen.getByPlaceholderText('Type a command or search...')).toBeInTheDocument()
    })
  })

  it('closes when escape is pressed', async () => {
    render(<CommandPalette />)

    // Open it first
    fireEvent.keyDown(document, { key: 'k', metaKey: true })
    await waitFor(() => {
        expect(screen.getByPlaceholderText('Type a command or search...')).toBeInTheDocument()
    })

    // Close it by pressing Escape on the dialog content mostly
    // We need to target the element that captures the focus or the document
    const input = screen.getByPlaceholderText('Type a command or search...')
    fireEvent.keyDown(input, { key: 'Escape', code: 'Escape' })

    // Radix Dialog might take a bit to close or unmount
    await waitFor(() => {
         expect(screen.queryByPlaceholderText('Type a command or search...')).not.toBeInTheDocument()
    })
  })
})
