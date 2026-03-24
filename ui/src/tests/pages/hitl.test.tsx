import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import HITLPage from '@/app/hitl/page';
import { BrowserRouter } from 'react-router-dom';

describe('HITL Approval Page', () => {
    it('renders the HITL approval page correctly', () => {
        render(
            <BrowserRouter>
                <HITLPage />
            </BrowserRouter>
        );

        expect(screen.getByText('HITL Approvals')).toBeDefined();
        expect(screen.getByText('Pending Approvals')).toBeDefined();
        expect(screen.getByText('Approved Actions')).toBeDefined();
        expect(screen.getByText('Denied Actions')).toBeDefined();
        expect(screen.getByText('Approval Queue')).toBeDefined();
    });
});
