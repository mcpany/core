import { render, screen, fireEvent } from '@testing-library/react'
import { GlobalSearch } from '../../components/global-search'
import { useRouter } from 'next/navigation'
import { useTheme } from 'next-themes'

// Mock dependencies
jest.mock('next/navigation', () => ({
  useRouter: jest.fn(),
}))
jest.mock('next-themes', () => ({
  useTheme: jest.fn(),
}))
// Mock ResizeObserver
global.ResizeObserver = jest.fn().mockImplementation(() => ({
  observe: jest.fn(),
  unobserve: jest.fn(),
  disconnect: jest.fn(),
}))

// Mock scrollIntoView
window.HTMLElement.prototype.scrollIntoView = jest.fn();

describe('GlobalSearch', () => {
  const mockPush = jest.fn()
  const mockSetTheme = jest.fn()

  beforeEach(() => {
    (useRouter as jest.Mock).mockReturnValue({ push: mockPush });
    (useTheme as jest.Mock).mockReturnValue({ setTheme: mockSetTheme });
  })

  afterEach(() => {
    jest.clearAllMocks()
  })

  it('renders the search button', () => {
    render(<GlobalSearch />)
    expect(screen.getByText(/Search feature/i)).toBeInTheDocument()
  })

  it('opens command palette when button is clicked', () => {
    render(<GlobalSearch />)
    const button = screen.getByText(/Search feature/i)
    fireEvent.click(button)
    // shadcn/cmdk renders input with a placeholder
    expect(screen.getByPlaceholderText(/Type a command/i)).toBeInTheDocument()
  })

  it('navigates when an item is selected', () => {
    render(<GlobalSearch />)
    // Open dialog
    fireEvent.click(screen.getByText(/Search feature/i))

    // Find Dashboard item and click it
    const dashboardItem = screen.getByText('Dashboard')
    fireEvent.click(dashboardItem)

    expect(mockPush).toHaveBeenCalledWith('/')
  })

  it('changes theme when a theme item is selected', () => {
    render(<GlobalSearch />)
    fireEvent.click(screen.getByText(/Search feature/i))

    const darkThemeItem = screen.getByText('Dark')
    fireEvent.click(darkThemeItem)

    expect(mockSetTheme).toHaveBeenCalledWith('dark')
  })
})
