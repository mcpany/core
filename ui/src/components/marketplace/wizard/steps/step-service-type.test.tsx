
import { render, screen, fireEvent } from '@testing-library/react';
import { StepServiceType } from './step-service-type';
import { WizardProvider } from '../wizard-context';
import { SERVICE_REGISTRY } from '@/lib/service-registry';
import { describe, it, expect } from 'vitest';

describe('StepServiceType', () => {
    it('renders registry templates', () => {
        render(
            <WizardProvider>
                <StepServiceType />
            </WizardProvider>
        );

        // Check if manual template is there (it appears in trigger and card, so multiple times)
        const manuals = screen.getAllByText('Manual / Custom');
        expect(manuals.length).toBeGreaterThan(0);

        // Check if a registry template is there (e.g. PostgreSQL)
        // Note: The select trigger shows "Select a template" initially or the selected one.
        // We need to click the trigger to see options.
        // The trigger is a button with role combobox
        const trigger = screen.getByRole('combobox');
        fireEvent.click(trigger);

        // Now options should be visible
        // We check for the name of the first item in registry
        const firstRegistryItem = SERVICE_REGISTRY[0];

        // Use findAllByText because it might be async or inside a portal
        screen.findAllByText(firstRegistryItem.name).then(items => {
             expect(items.length).toBeGreaterThan(0);
        });
    });
});
