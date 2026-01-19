import { render, screen, fireEvent } from '@testing-library/react';
import { ChatMessage, Message } from '@/components/playground/pro/chat-message';
import { vi, describe, it, expect } from 'vitest';

describe('ChatMessage', () => {
  it('renders user message correctly', () => {
    const message: Message = {
      id: '1',
      type: 'user',
      content: 'Hello world',
      timestamp: new Date(),
    };
    render(<ChatMessage message={message} />);
    expect(screen.getByText('Hello world')).toBeInTheDocument();
  });

  it('renders tool-call message correctly with replay button', () => {
    const message: Message = {
      id: '2',
      type: 'tool-call',
      toolName: 'my_tool',
      toolArgs: { arg: 'val' },
      timestamp: new Date(),
    };
    const onReplay = vi.fn();

    render(<ChatMessage message={message} onReplay={onReplay} />);

    expect(screen.getByText('my_tool')).toBeInTheDocument();

    const replayButton = screen.getByLabelText('Load into console');
    fireEvent.click(replayButton);

    expect(onReplay).toHaveBeenCalledWith('my_tool', { arg: 'val' });
  });

  it('does not render replay button if onReplay is not provided', () => {
     const message: Message = {
      id: '3',
      type: 'tool-call',
      toolName: 'my_tool',
      toolArgs: { arg: 'val' },
      timestamp: new Date(),
    };
    render(<ChatMessage message={message} />);

    expect(screen.queryByLabelText('Load into console')).not.toBeInTheDocument();
  });
});
