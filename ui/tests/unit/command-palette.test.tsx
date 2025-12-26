
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
    })
}))

// Mock Dialog to render children immediately to simplify testing
// We don't need to test Radix UI's Dialog implementation, just our logic
jest.mock('../../src/components/ui/dialog', () => ({
    Dialog: ({ children, open, onOpenChange }: any) => {
        if (!open) return null;
        return <div data-testid="command-dialog">{children}</div>
    },
    DialogContent: ({ children }: any) => <div>{children}</div>
}));

jest.mock('../../src/components/ui/command', () => {
    const React = require('react');
    return {
        Command: ({ children }: any) => <div data-testid="command-root">{children}</div>,
        CommandDialog: ({ children, open, onOpenChange }: any) => {
             if (!open) return null;
             return <div data-testid="command-dialog">{children}</div>
        },
        CommandInput: ({ placeholder, ...props }: any) => <input placeholder={placeholder} {...props} />,
        CommandList: ({ children }: any) => <div>{children}</div>,
        CommandEmpty: ({ children }: any) => <div>{children}</div>,
        CommandGroup: ({ children, heading }: any) => (
            <div>
                <div>{heading}</div>
                {children}
            </div>
        ),
        CommandItem: ({ children, onSelect }: any) => (
            <div onClick={onSelect} data-testid="command-item">
                {children}
            </div>
        ),
        CommandSeparator: () => <hr />,
        CommandShortcut: ({ children }: any) => <span>{children}</span>,
    }
});


describe('CommandPalette', () => {
  it('does not render initially', () => {
    render(<CommandPalette />)
    expect(screen.queryByTestId('command-dialog')).not.toBeInTheDocument()
  })

  it('opens when Cmd+K is pressed', async () => {
    render(<CommandPalette />)

    fireEvent.keyDown(document, { key: 'k', metaKey: true })

    await waitFor(() => {
        expect(screen.getByTestId('command-dialog')).toBeInTheDocument()
    })
  })

  it('opens when Ctrl+K is pressed', async () => {
    render(<CommandPalette />)

    fireEvent.keyDown(document, { key: 'k', ctrlKey: true })

    await waitFor(() => {
        expect(screen.getByTestId('command-dialog')).toBeInTheDocument()
    })
  })
})
