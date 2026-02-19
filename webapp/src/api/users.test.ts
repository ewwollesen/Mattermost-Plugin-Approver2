/**
 * User API Tests
 * Story 11.2: User Selector Component - Task 2
 */

import {searchUsers, getDisplayName, UserOption} from './users';

// Mock fetch globally
const mockFetch = jest.fn();
global.fetch = mockFetch;

describe('User API (Task 2)', () => {
    beforeEach(() => {
        jest.clearAllMocks();
    });

    describe('searchUsers', () => {
        const mockUsers = [
            {id: 'user1', username: 'john.doe', first_name: 'John', last_name: 'Doe', nickname: ''},
            {id: 'user2', username: 'jane.smith', first_name: 'Jane', last_name: 'Smith', nickname: ''},
        ];

        it('returns empty array for search term less than 2 characters', async () => {
            const result = await searchUsers('j');
            expect(result).toEqual([]);
            expect(mockFetch).not.toHaveBeenCalled();
        });

        it('calls autocomplete API with correct URL', async () => {
            mockFetch.mockResolvedValueOnce({
                ok: true,
                json: () => Promise.resolve({users: mockUsers}),
            });

            await searchUsers('john');

            expect(mockFetch).toHaveBeenCalledWith(
                '/api/v4/users/autocomplete?term=john',
                expect.objectContaining({
                    method: 'GET',
                    headers: {'Content-Type': 'application/json'},
                })
            );
        });

        it('transforms API response to UserOption format', async () => {
            mockFetch.mockResolvedValueOnce({
                ok: true,
                json: () => Promise.resolve({users: mockUsers}),
            });

            const result = await searchUsers('john');

            expect(result).toHaveLength(2);
            expect(result[0]).toEqual({
                id: 'user1',
                username: 'john.doe',
                displayName: 'John Doe',
                avatarUrl: expect.stringContaining('/api/v4/users/user1/image'),
            });
        });

        it('returns empty array on API error', async () => {
            mockFetch.mockResolvedValueOnce({
                ok: false,
                status: 500,
            });

            const consoleSpy = jest.spyOn(console, 'error').mockImplementation();
            const result = await searchUsers('john');

            expect(result).toEqual([]);
            expect(consoleSpy).toHaveBeenCalled();
            consoleSpy.mockRestore();
        });

        it('returns empty array on network error', async () => {
            mockFetch.mockRejectedValueOnce(new Error('Network error'));

            const consoleSpy = jest.spyOn(console, 'error').mockImplementation();
            const result = await searchUsers('john');

            expect(result).toEqual([]);
            expect(consoleSpy).toHaveBeenCalled();
            consoleSpy.mockRestore();
        });

        it('URL-encodes special characters in search term', async () => {
            mockFetch.mockResolvedValueOnce({
                ok: true,
                json: () => Promise.resolve({users: []}),
            });

            await searchUsers('john doe');

            expect(mockFetch).toHaveBeenCalledWith(
                '/api/v4/users/autocomplete?term=john%20doe',
                expect.any(Object)
            );
        });

        it('handles empty users array in response', async () => {
            mockFetch.mockResolvedValueOnce({
                ok: true,
                json: () => Promise.resolve({users: []}),
            });

            const result = await searchUsers('xyz');

            expect(result).toEqual([]);
        });
    });

    describe('getDisplayName', () => {
        it('returns nickname if available', () => {
            const user = {
                id: '1',
                username: 'john',
                first_name: 'John',
                last_name: 'Doe',
                nickname: 'Johnny',
            };
            expect(getDisplayName(user)).toBe('Johnny');
        });

        it('returns first and last name if no nickname', () => {
            const user = {
                id: '1',
                username: 'john',
                first_name: 'John',
                last_name: 'Doe',
                nickname: '',
            };
            expect(getDisplayName(user)).toBe('John Doe');
        });

        it('returns first name only if no last name', () => {
            const user = {
                id: '1',
                username: 'john',
                first_name: 'John',
                last_name: '',
                nickname: '',
            };
            expect(getDisplayName(user)).toBe('John');
        });

        it('returns username if no name fields', () => {
            const user = {
                id: '1',
                username: 'john',
                first_name: '',
                last_name: '',
                nickname: '',
            };
            expect(getDisplayName(user)).toBe('john');
        });
    });
});
