
import { render, screen } from '@testing-library/react';
import { ServiceList } from '@/components/services/service-list';
import { UpstreamServiceConfig } from '@/lib/client';

describe('ServiceList', () => {
    const mockServices: UpstreamServiceConfig[] = [
        {
            name: 'test-service-1',
            version: '1.0.0',
            disable: false,
            http_service: { address: 'http://localhost:8080' }
        },
        {
            name: 'test-service-2',
            version: '2.0.0',
            disable: true,
            grpc_service: { address: 'localhost:9090' }
        }
    ];

    it('renders services in a table for desktop', () => {
        render(<ServiceList services={mockServices} />);

        // Should exist in the document (twice now, desktop and mobile)
        const items = screen.getAllByText('test-service-1');
        expect(items.length).toBeGreaterThanOrEqual(1);
    });

    it('applies responsive classes to table', () => {
        const { container } = render(<ServiceList services={mockServices} />);

        // Desktop container
        const desktopWrapper = container.querySelector('.hidden.md\\:block');
        expect(desktopWrapper).toBeInTheDocument();
    });

    it('renders services as cards for mobile', () => {
        const { container } = render(<ServiceList services={mockServices} />);

        // Mobile container
        // Note: class selector with colon needs escaping in JS querySelector if strictly following CSS rules,
        // but simple class string matching might be easier if we just look for the div.
        const mobileView = container.querySelector('.md\\:hidden.grid');
        expect(mobileView).toBeInTheDocument();

        if (mobileView) {
             expect(mobileView).toHaveTextContent('test-service-1');
        }
    });
});
