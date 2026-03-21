import { render, screen } from '@testing-library/react';
import { describe, it, expect } from 'vitest';
import { RichResultViewer, flattenObject } from './rich-result-viewer';

describe('flattenObject', () => {
    it('flattens deeply nested object', () => {
        const obj = {
            user: {
                profile: {
                    name: 'Alice',
                    age: 28
                },
                role: 'admin'
            },
            metadata: {
                theme: 'dark'
            }
        };

        const result = flattenObject(obj);

        expect(result).toEqual({
            'user.profile.name': 'Alice',
            'user.profile.age': 28,
            'user.role': 'admin',
            'metadata.theme': 'dark'
        });
    });

    it('stringifies nested arrays', () => {
        const obj = {
            contacts: [
                { type: 'email', value: 'a@b.com' },
                { type: 'phone', value: '123' }
            ]
        };

        const result = flattenObject(obj);

        expect(result).toEqual({
            'contacts': 'type: a@b.com, type: 123'
        });
    });

    it('handles primitive arrays correctly', () => {
        const obj = {
            tags: ['foo', 'bar']
        };

        const result = flattenObject(obj);

        expect(result).toEqual({
            'tags': 'foo, bar'
        });
    });
});

describe('RichResultViewer Component', () => {
    it('renders flattened nested array of objects as table', () => {
        const complexData = [
            {
                user: {
                    profile: { name: 'Alice' },
                    role: 'admin'
                },
                tags: ['a', 'b']
            },
            {
                user: {
                    profile: { name: 'Bob' },
                    role: 'user'
                },
                tags: ['c', 'd']
            }
        ];

        render(<RichResultViewer result={complexData} />);

        // Headers
        expect(screen.getByText('USER.PROFILE.NAME')).toBeInTheDocument();
        expect(screen.getByText('USER.ROLE')).toBeInTheDocument();
        expect(screen.getByText('TAGS')).toBeInTheDocument();

        // Data cells
        expect(screen.getByText('Alice')).toBeInTheDocument();
        expect(screen.getByText('admin')).toBeInTheDocument();
        expect(screen.getByText('a, b')).toBeInTheDocument();

        expect(screen.getByText('Bob')).toBeInTheDocument();
        expect(screen.getByText('user')).toBeInTheDocument();
        expect(screen.getByText('c, d')).toBeInTheDocument();
    });
});
