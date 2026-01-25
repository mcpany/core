
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { ResourceExplorer } from '@/components/resources/resource-explorer';
import { apiClient } from '@/lib/client';
import { vi } from 'vitest';

// Mock dependencies
vi.mock('@/lib/client', () => ({
    apiClient: {
        listResources: vi.fn(),
        readResource: vi.fn(),
    },
}));

vi.mock('@/hooks/use-toast', () => ({
    useToast: () => ({ toast: vi.fn() }),
}));

// Mock URL.createObjectURL
global.URL.createObjectURL = vi.fn(() => 'blob:test');
global.URL.revokeObjectURL = vi.fn();

describe('ResourceExplorer', () => {
    const mockResources = [
        { uri: 'file:///test.txt', name: 'test.txt', mimeType: 'text/plain' },
    ];

    beforeEach(() => {
        vi.clearAllMocks();
        (apiClient.listResources as unknown as ReturnType<typeof vi.fn>).mockResolvedValue({ resources: mockResources });
    });

    it('sets DownloadURL on drag start when content is loaded', async () => {
        (apiClient.readResource as unknown as ReturnType<typeof vi.fn>).mockResolvedValue({
            contents: [{ uri: 'file:///test.txt', mimeType: 'text/plain', text: 'hello' }],
        });

        render(<ResourceExplorer initialResources={mockResources} />);

        // Click to load content
        const item = screen.getByText('test.txt');
        fireEvent.click(item);

        await waitFor(() => expect(apiClient.readResource).toHaveBeenCalledWith('file:///test.txt'));

        // Wait for content to render (ResourceViewer text)
        await waitFor(() => expect(screen.getByTitle('Drag to desktop to save')).toBeInTheDocument());

        // Find drag handle (File icon)
        const dragHandle = screen.getByTitle('Drag to desktop to save');
        expect(dragHandle).toBeInTheDocument();

        // Simulate drag start
        const dataTransfer = { setData: vi.fn() };
        fireEvent.dragStart(dragHandle, { dataTransfer });

        expect(dataTransfer.setData).toHaveBeenCalledWith('DownloadURL', 'text/plain:test.txt:blob:test');
        expect(dataTransfer.setData).toHaveBeenCalledWith('text/plain', 'file:///test.txt');
    });

    it('handleDownload fetches content if missing', async () => {
        (apiClient.readResource as unknown as ReturnType<typeof vi.fn>).mockResolvedValue({
            contents: [{ uri: 'file:///test.txt', mimeType: 'text/plain', text: 'hello' }],
        });

        render(<ResourceExplorer initialResources={mockResources} />);

        // Mock anchor click
        const clickMock = vi.fn();
        const originalCreateElement = document.createElement;
        const createElementSpy = vi.spyOn(document, 'createElement').mockImplementation((tagName, options) => {
            const el = originalCreateElement.call(document, tagName, options);
            if (tagName === 'a') {
                el.click = clickMock;
            }
            return el;
        });

        // Don't select (click). Just right click (context menu).
        const item = screen.getByText('test.txt');
        fireEvent.contextMenu(item);

        // ContextMenu content renders in a portal. We need to wait for it.
        const downloadMenuItem = await screen.findByText('Download');
        fireEvent.click(downloadMenuItem);

        await waitFor(() => expect(apiClient.readResource).toHaveBeenCalledWith('file:///test.txt'));
        await waitFor(() => expect(clickMock).toHaveBeenCalled());

        createElementSpy.mockRestore();
    });
});
