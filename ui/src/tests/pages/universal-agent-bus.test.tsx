import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import UniversalAgentBusPage from '@/app/universal-agent-bus/page';
import { BrowserRouter } from 'react-router-dom';

describe('Universal Agent Bus Page', () => {
    it('renders the universal agent bus page correctly', () => {
        render(
            <BrowserRouter>
                <UniversalAgentBusPage />
            </BrowserRouter>
        );

        expect(screen.getByText('Universal Agent Bus')).toBeDefined();
        expect(screen.getByText('Recursive Context Dashboard')).toBeDefined();
        expect(screen.getByText('Multi-Agent Session Timeline')).toBeDefined();
        expect(screen.getByText('Unified Discovery Manager')).toBeDefined();
        expect(screen.getByText('Lazy-MCP Tool Search')).toBeDefined();
        expect(screen.getByText('Agent Chain Tracer (A2A)')).toBeDefined();
    });
});
