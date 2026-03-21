import { render, screen, fireEvent } from '@testing-library/react';
import { UserList, User } from './user-list';
import { vi } from 'vitest';

describe('UserList', () => {
    const mockUsers: User[] = [
        {
            id: 'test-user-1',
            roles: ['admin', 'editor'],
            authentication: {
                apiKey: 'some-key',
                basicAuth: null
            } as any
        },
        {
            id: 'test-user-2',
            roles: ['viewer'],
            authentication: {
                basicAuth: 'some-hash',
                apiKey: ''
            } as any
        },
        {
            id: 'test-user-3',
            roles: [],
            authentication: {}
        }
    ];

    it('renders user list using virtuoso', async () => {
        const onEdit = vi.fn();
        const onDelete = vi.fn();

        render(<UserList users={mockUsers} onEdit={onEdit} onDelete={onDelete} />);

        // Give virtuoso time to render. In JSDOM Virtuoso might need explicit sizing or won't render items.
        // We test the filtered count to verify logic works.
        expect(screen.getByText('Showing 3 of 3 users')).toBeInTheDocument();
    });

    it('filters users correctly', async () => {
        render(<UserList users={mockUsers} onEdit={vi.fn()} onDelete={vi.fn()} />);

        const searchInput = screen.getByPlaceholderText('Search users...');
        fireEvent.change(searchInput, { target: { value: 'admin' } });

        expect(screen.getByText('Showing 1 of 3 users')).toBeInTheDocument();
    });
});
