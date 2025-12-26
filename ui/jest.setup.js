import '@testing-library/jest-dom'

global.ResizeObserver = class ResizeObserver {
  observe() {}
  unobserve() {}
  disconnect() {}
}

window.HTMLElement.prototype.scrollIntoView = function() {};
window.HTMLElement.prototype.hasPointerCapture = function() { return false; };
window.HTMLElement.prototype.setPointerCapture = function() {};
window.HTMLElement.prototype.releasePointerCapture = function() {};

// Mocking Dialog which uses portals
jest.mock('@radix-ui/react-dialog', () => {
    return {
        ...jest.requireActual('@radix-ui/react-dialog'),
        Root: ({ open, children, onOpenChange }) => {
            if (!open) return null;
            return <div data-testid="command-dialog">{children}</div>
        },
        Portal: ({ children }) => <div>{children}</div>,
        Overlay: ({ children }) => <div>{children}</div>,
        Content: ({ children, className }) => <div className={className}>{children}</div>,
        Close: ({ children }) => <div>{children}</div>,
    }
})

// Also mock cmdk to avoid complex issues in jsdom environment if necessary,
// but let's try with just ResizeObserver first, as cmdk relies on it.
