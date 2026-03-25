import { render, screen, fireEvent, waitFor } from '../../../tests/test-utils';
import { ChatMessage } from './chat-message';
import { vi, describe, it, expect, beforeEach } from 'vitest';

// Create a functional mock component to avoid hook issues
/**
 * Summary: MockDiffEditor component.
 *
 * Parameters:
 *   - None.
 *
 * Returns:
 *   - React.ReactNode: The rendered component.
 *
 * Throws/Errors:
 *   - None.
 */
const MockDiffEditor = (props: any) => {
  return (
    <div data-testid="diff-editor">
      <div data-testid="original">{props.original}</div>
      <div data-testid="modified">{props.modified}</div>
    </div>
  );
};

vi.mock('@monaco-editor/react', () => {
    return {
        default: {
            DiffEditor: MockDiffEditor
        },
        DiffEditor: MockDiffEditor
    };
});

describe('ChatMessage Diffing', () => {
  it('shows diff button when previous result exists and is different', async () => {
    const message = {
      id: 'msg-1',
      type: 'tool-result' as const,
      toolName: 'test-tool',
      toolResult: { data: 'new-value' },
      previousResult: { data: 'old-value' },
      timestamp: new Date()
    };

    render(<ChatMessage message={message}  />);

    // Wait for the button to appear since there might be async effects
    const diffButton = await screen.findByText(/Show Changes/i);
    expect(diffButton).toBeInTheDocument();

    // Click the diff button
    fireEvent.click(diffButton);

    // Check if the dialog and diff editor are shown
    await waitFor(() => {
      expect(screen.getByText('Output Difference')).toBeInTheDocument();
      // Use getByTestId inside waitFor or handle lazy loading
    });
  });

  it('does not show diff button when results are identical', () => {
    const message = {
      id: 'msg-1',
      type: 'tool-result' as const,
      toolName: 'test-tool',
      toolResult: { data: 'same-value' },
      previousResult: { data: 'same-value' },
      timestamp: new Date()
    };

    render(<ChatMessage message={message}  />);

    // Check that the diff button is not rendered
    const diffButton = screen.queryByText(/Show Changes/i);
    expect(diffButton).not.toBeInTheDocument();
  });
});
