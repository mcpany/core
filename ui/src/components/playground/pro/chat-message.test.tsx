/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen } from '@testing-library/react';
import { ChatMessage, Message } from './chat-message';
import { describe, it, expect, vi } from 'vitest';

// Mock dependencies
// Mocking react-syntax-highlighter to avoid heavy rendering and potential JSDOM issues
vi.mock('react-syntax-highlighter', () => ({
    Prism: ({ children, language }: any) => (
        <pre data-testid="code-block" data-language={language}>
            {children}
        </pre>
    ),
}));

vi.mock('@monaco-editor/react', () => ({
    DiffEditor: () => <div>DiffEditor Mock</div>,
}));

vi.mock('next-themes', () => ({
    useTheme: () => ({ theme: 'light' }),
}));

describe('ChatMessage', () => {
    it('renders user message with markdown', () => {
        const message: Message = {
            id: '1',
            type: 'user',
            content: '**Hello** world',
            timestamp: new Date(),
        };

        const { container } = render(<ChatMessage message={message} />);

        // Markdown should render **Hello** as strong/bold
        const strongElement = container.querySelector('strong');
        expect(strongElement).toBeInTheDocument();
        expect(strongElement).toHaveTextContent('Hello');
        expect(screen.getByText('world')).toBeInTheDocument();
    });

    it('renders assistant message with code block', () => {
        const message: Message = {
            id: '2',
            type: 'assistant',
            content: 'Here is code:\n```javascript\nconst a = 1;\n```',
            timestamp: new Date(),
        };

        const { container } = render(<ChatMessage message={message} />);

        expect(screen.getByText('Here is code:')).toBeInTheDocument();

        // Check for our mocked code block
        const codeBlock = screen.getByTestId('code-block');
        expect(codeBlock).toBeInTheDocument();
        expect(codeBlock).toHaveAttribute('data-language', 'javascript');
        expect(codeBlock).toHaveTextContent('const a = 1;');
    });

    it('renders assistant message with table', () => {
        const message: Message = {
            id: '3',
            type: 'assistant',
            content: '| Header 1 | Header 2 |\n| --- | --- |\n| Cell 1 | Cell 2 |',
            timestamp: new Date(),
        };

        const { container } = render(<ChatMessage message={message} />);

        expect(container.querySelector('table')).toBeInTheDocument();
        expect(screen.getByText('Header 1')).toBeInTheDocument();
        expect(screen.getByText('Cell 1')).toBeInTheDocument();
    });
});
