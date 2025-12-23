
import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { GlobalSearch } from '../../src/components/global-search';

// Mock useRouter
jest.mock('next/navigation', () => ({
  useRouter() {
    return {
      push: jest.fn(),
    };
  },
}));

// Mock Lucide icons
jest.mock('lucide-react', () => ({
    Calculator: () => <span>CalculatorIcon</span>,
    Calendar: () => <span>CalendarIcon</span>,
    CreditCard: () => <span>CreditCardIcon</span>,
    Settings: () => <span>SettingsIcon</span>,
    Smile: () => <span>SmileIcon</span>,
    User: () => <span>UserIcon</span>,
    LayoutDashboard: () => <span>LayoutDashboardIcon</span>,
    Server: () => <span>ServerIcon</span>,
    Wrench: () => <span>WrenchIcon</span>,
    Search: () => <span>SearchIcon</span>,
    ExternalLink: () => <span>ExternalLinkIcon</span>,
    Laptop: () => <span>LaptopIcon</span>,
}));

// Mock Dialog
jest.mock('@/components/ui/dialog', () => ({
    Dialog: ({ children, open, onOpenChange }: any) => {
        return open ? <div>{children}</div> : null;
    },
    DialogContent: ({ children }: any) => <div>{children}</div>,
}));

// Mock Command
jest.mock('@/components/ui/command', () => ({
    Command: ({ children }: any) => <div>{children}</div>,
    CommandDialog: ({ children, open, onOpenChange }: any) => open ? <div>{children}</div> : null,
    CommandInput: ({ placeholder }: any) => <input placeholder={placeholder} />,
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
}));

describe('GlobalSearch', () => {
    it('renders the trigger button', () => {
        render(<GlobalSearch />);
        // There are two buttons, one for desktop and one for mobile
        const buttons = screen.getAllByRole('button');
        expect(buttons.length).toBeGreaterThan(0);
    });

    it('opens the search dialog when button is clicked', () => {
        render(<GlobalSearch />);
        const button = screen.getAllByRole('button')[0]; // Pick the first one
        fireEvent.click(button);

        expect(screen.getByPlaceholderText('Type a command or search...')).toBeInTheDocument();
    });

    it('renders suggestions', () => {
        render(<GlobalSearch />);
        const button = screen.getAllByRole('button')[0];
        fireEvent.click(button);

        expect(screen.getByText('Suggestions')).toBeInTheDocument();
        expect(screen.getByText('Dashboard')).toBeInTheDocument();
        expect(screen.getByText('Services')).toBeInTheDocument();
    });
});
