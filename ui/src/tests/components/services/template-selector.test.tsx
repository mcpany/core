import { render, screen, fireEvent } from '@testing-library/react';
import { TemplateSelector } from '@/components/services/template-selector';
import { SERVICE_TEMPLATES } from '@/data/service-templates';
import '@testing-library/jest-dom';
import { vi } from 'vitest';

describe('TemplateSelector', () => {
    it('renders all templates', () => {
        const mockSelect = vi.fn();
        const mockCancel = vi.fn();

        render(<TemplateSelector onSelect={mockSelect} onCancel={mockCancel} />);

        // Check for "Custom Service" card
        expect(screen.getByText('Custom Service')).toBeInTheDocument();

        // Check for each template
        SERVICE_TEMPLATES.forEach(template => {
            expect(screen.getByText(template.name)).toBeInTheDocument();
        });
    });

    it('calls onSelect when a template is clicked', () => {
        const mockSelect = vi.fn();
        const mockCancel = vi.fn();

        render(<TemplateSelector onSelect={mockSelect} onCancel={mockCancel} />);

        // Click the first template (Generic HTTP API)
        const templateName = SERVICE_TEMPLATES[0].name;
        fireEvent.click(screen.getByText(templateName));

        expect(mockSelect).toHaveBeenCalledWith(SERVICE_TEMPLATES[0]);
    });

    it('calls onSelect with custom template when Custom Service is clicked', () => {
        const mockSelect = vi.fn();
        const mockCancel = vi.fn();

        render(<TemplateSelector onSelect={mockSelect} onCancel={mockCancel} />);

        fireEvent.click(screen.getByText('Custom Service'));

        expect(mockSelect).toHaveBeenCalledWith(expect.objectContaining({
            id: 'custom',
            name: 'Custom Service'
        }));
    });

    it('calls onCancel when Cancel button is clicked', () => {
        const mockSelect = vi.fn();
        const mockCancel = vi.fn();

        render(<TemplateSelector onSelect={mockSelect} onCancel={mockCancel} />);

        fireEvent.click(screen.getByText('Cancel'));

        expect(mockCancel).toHaveBeenCalled();
    });
});
