
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { CommandPalette } from '../../src/components/command-palette'
import { useRouter } from 'next/navigation'

// Mocking Next.js router
jest.mock('next/navigation', () => ({
  useRouter: jest.fn(),
}))

// Mocking useTheme
jest.mock('next-themes', () => ({
    useTheme: () => ({
        setTheme: jest.fn(),
        theme: 'light'
    }),
    ThemeProvider: ({children}: {children: React.ReactNode}) => <div>{children}</div>
}))


describe('CommandPalette Integration', () => {
  it('navigates to dashboard when Dashboard item is clicked', async () => {
    const push = jest.fn()
    ;(useRouter as jest.Mock).mockReturnValue({ push })

    render(<CommandPalette />)

    // Open
    fireEvent.keyDown(document, { key: 'k', metaKey: true })

    await waitFor(() => {
        expect(screen.getByPlaceholderText('Type a command or search...')).toBeInTheDocument()
    })

    // Click Dashboard
    const dashboardItem = screen.getByText('Dashboard')
    fireEvent.click(dashboardItem)

    // Check if router.push was called
    expect(push).toHaveBeenCalledWith('/')
  })

   it('navigates to settings when Settings item is clicked', async () => {
    const push = jest.fn()
    ;(useRouter as jest.Mock).mockReturnValue({ push })

    render(<CommandPalette />)

    // Open
    fireEvent.keyDown(document, { key: 'k', metaKey: true })

     await waitFor(() => {
        expect(screen.getByPlaceholderText('Type a command or search...')).toBeInTheDocument()
    })

    // Click Settings
    const settingsItem = screen.getByText('Settings')
    fireEvent.click(settingsItem)

    // Check if router.push was called
    expect(push).toHaveBeenCalledWith('/settings')
  })
})
