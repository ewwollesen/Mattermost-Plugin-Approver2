/**
 * UserSelector Component Tests
 * Story 11.2: User Selector Component
 */

import React from 'react';
import {render, screen, fireEvent, waitFor, act} from '@testing-library/react';
import UserSelector from './UserSelector';

// Mock fetch globally
const mockFetch = jest.fn();
global.fetch = mockFetch;

// Mock user data
const mockUsers = [
    {id: 'user1', username: 'john.doe', first_name: 'John', last_name: 'Doe', nickname: ''},
    {id: 'user2', username: 'jane.smith', first_name: 'Jane', last_name: 'Smith', nickname: ''},
    {id: 'user3', username: 'bob.wilson', first_name: 'Bob', last_name: 'Wilson', nickname: 'Bobby'},
];

const mockApiResponse = {
    users: mockUsers,
};

describe('UserSelector Component', () => {
    const defaultProps = {
        value: '',
        onChange: jest.fn(),
    };

    beforeEach(() => {
        jest.clearAllMocks();
        jest.useFakeTimers();
        mockFetch.mockResolvedValue({
            ok: true,
            json: () => Promise.resolve(mockApiResponse),
        });
    });

    afterEach(() => {
        jest.useRealTimers();
    });

    describe('Task 1: Component Structure (AC: 1, 3)', () => {
        it('renders with default props', () => {
            render(<UserSelector {...defaultProps} />);
            expect(screen.getByTestId('user-selector')).toBeInTheDocument();
        });

        it('renders search input', () => {
            render(<UserSelector {...defaultProps} placeholder="Search users..." />);
            const input = screen.getByPlaceholderText('Search users...');
            expect(input).toBeInTheDocument();
        });

        it('renders label when provided', () => {
            render(<UserSelector {...defaultProps} label="Select Approver" />);
            expect(screen.getByText('Select Approver')).toBeInTheDocument();
        });

        it('renders required indicator when label provided', () => {
            render(<UserSelector {...defaultProps} label="Select Approver" />);
            expect(screen.getByText('*')).toBeInTheDocument();
        });

        it('renders search icon', () => {
            render(<UserSelector {...defaultProps} />);
            expect(screen.getByTestId('search-icon')).toBeInTheDocument();
        });

        it('applies disabled state', () => {
            render(<UserSelector {...defaultProps} disabled={true} />);
            const input = screen.getByRole('combobox');
            expect(input).toBeDisabled();
        });

        it('uses Mattermost CSS variable styling', () => {
            render(<UserSelector {...defaultProps} />);
            const container = screen.getByTestId('user-selector');
            expect(container).toBeInTheDocument();
            // Verify inline styles use CSS variables
            const input = screen.getByRole('combobox');
            expect(input).toHaveStyle({
                backgroundColor: 'var(--center-channel-bg, #ffffff)',
            });
        });
    });

    describe('Task 3: Dropdown Results List (AC: 1, 3)', () => {
        it('shows dropdown when user types 2+ characters', async () => {
            render(<UserSelector {...defaultProps} />);
            const input = screen.getByRole('combobox');

            fireEvent.change(input, {target: {value: 'jo'}});

            // Wait for debounce
            act(() => {
                jest.advanceTimersByTime(300);
            });

            await waitFor(() => {
                expect(screen.getByTestId('user-dropdown')).toBeInTheDocument();
            });
        });

        it('does not show dropdown for single character input', async () => {
            render(<UserSelector {...defaultProps} />);
            const input = screen.getByRole('combobox');

            fireEvent.change(input, {target: {value: 'j'}});

            act(() => {
                jest.advanceTimersByTime(300);
            });

            expect(screen.queryByTestId('user-dropdown')).not.toBeInTheDocument();
        });

        it('displays user results with avatar, display name, and username', async () => {
            render(<UserSelector {...defaultProps} />);
            const input = screen.getByRole('combobox');

            fireEvent.change(input, {target: {value: 'john'}});

            act(() => {
                jest.advanceTimersByTime(300);
            });

            await waitFor(() => {
                expect(screen.getByTestId('user-dropdown')).toBeInTheDocument();
            });

            // Check for user option with display name and username
            expect(screen.getByText('John Doe')).toBeInTheDocument();
            expect(screen.getByText('@john.doe')).toBeInTheDocument();
            // Check for avatar image
            expect(screen.getByTestId('user-avatar-user1')).toBeInTheDocument();
        });

        it('shows loading spinner during API fetch', async () => {
            // Delay the API response
            mockFetch.mockImplementationOnce(() =>
                new Promise((resolve) =>
                    setTimeout(() =>
                        resolve({
                            ok: true,
                            json: () => Promise.resolve(mockApiResponse),
                        }),
                        500
                    )
                )
            );

            render(<UserSelector {...defaultProps} />);
            const input = screen.getByRole('combobox');

            fireEvent.change(input, {target: {value: 'john'}});

            act(() => {
                jest.advanceTimersByTime(300);
            });

            // Loading state should appear
            await waitFor(() => {
                expect(screen.getByTestId('loading-spinner')).toBeInTheDocument();
            });
        });

        it('shows "No users found" when search returns empty', async () => {
            mockFetch.mockResolvedValueOnce({
                ok: true,
                json: () => Promise.resolve({users: []}),
            });

            render(<UserSelector {...defaultProps} />);
            const input = screen.getByRole('combobox');

            fireEvent.change(input, {target: {value: 'xyz'}});

            act(() => {
                jest.advanceTimersByTime(300);
            });

            await waitFor(() => {
                expect(screen.getByText('No users found')).toBeInTheDocument();
            });
        });

        it('hides dropdown when clicking outside', async () => {
            render(
                <div>
                    <UserSelector {...defaultProps} />
                    <button data-testid="outside-element">Outside</button>
                </div>
            );
            const input = screen.getByRole('combobox');

            fireEvent.change(input, {target: {value: 'john'}});

            act(() => {
                jest.advanceTimersByTime(300);
            });

            await waitFor(() => {
                expect(screen.getByTestId('user-dropdown')).toBeInTheDocument();
            });

            // Click outside
            fireEvent.mouseDown(screen.getByTestId('outside-element'));

            await waitFor(() => {
                expect(screen.queryByTestId('user-dropdown')).not.toBeInTheDocument();
            });
        });

        it('hides dropdown when pressing Escape', async () => {
            render(<UserSelector {...defaultProps} />);
            const input = screen.getByRole('combobox');

            fireEvent.change(input, {target: {value: 'john'}});

            act(() => {
                jest.advanceTimersByTime(300);
            });

            await waitFor(() => {
                expect(screen.getByTestId('user-dropdown')).toBeInTheDocument();
            });

            fireEvent.keyDown(input, {key: 'Escape'});

            expect(screen.queryByTestId('user-dropdown')).not.toBeInTheDocument();
        });

        it('navigates through results with Arrow keys', async () => {
            render(<UserSelector {...defaultProps} />);
            const input = screen.getByRole('combobox');

            fireEvent.change(input, {target: {value: 'john'}});

            act(() => {
                jest.advanceTimersByTime(300);
            });

            await waitFor(() => {
                expect(screen.getByTestId('user-dropdown')).toBeInTheDocument();
            });

            // Press ArrowDown to highlight first item
            fireEvent.keyDown(input, {key: 'ArrowDown'});

            const firstOption = screen.getByTestId('user-option-user1');
            expect(firstOption).toHaveAttribute('data-highlighted', 'true');

            // Press ArrowDown to highlight second item
            fireEvent.keyDown(input, {key: 'ArrowDown'});

            const secondOption = screen.getByTestId('user-option-user2');
            expect(secondOption).toHaveAttribute('data-highlighted', 'true');
            expect(firstOption).toHaveAttribute('data-highlighted', 'false');

            // Press ArrowUp to go back
            fireEvent.keyDown(input, {key: 'ArrowUp'});
            expect(firstOption).toHaveAttribute('data-highlighted', 'true');
        });

        it('selects user with Enter key', async () => {
            const mockOnChange = jest.fn();
            render(<UserSelector {...defaultProps} onChange={mockOnChange} />);
            const input = screen.getByRole('combobox');

            fireEvent.change(input, {target: {value: 'john'}});

            act(() => {
                jest.advanceTimersByTime(300);
            });

            await waitFor(() => {
                expect(screen.getByTestId('user-dropdown')).toBeInTheDocument();
            });

            // Navigate and select
            fireEvent.keyDown(input, {key: 'ArrowDown'});
            fireEvent.keyDown(input, {key: 'Enter'});

            expect(mockOnChange).toHaveBeenCalledWith('user1');
        });

        it('selects user with mouse click', async () => {
            const mockOnChange = jest.fn();
            render(<UserSelector {...defaultProps} onChange={mockOnChange} />);
            const input = screen.getByRole('combobox');

            fireEvent.change(input, {target: {value: 'john'}});

            act(() => {
                jest.advanceTimersByTime(300);
            });

            await waitFor(() => {
                expect(screen.getByTestId('user-dropdown')).toBeInTheDocument();
            });

            // Click on user option
            fireEvent.click(screen.getByTestId('user-option-user1'));

            expect(mockOnChange).toHaveBeenCalledWith('user1');
        });

        it('excludes specified user IDs from results', async () => {
            render(<UserSelector {...defaultProps} excludeUserIds={['user1']} />);
            const input = screen.getByRole('combobox');

            fireEvent.change(input, {target: {value: 'john'}});

            act(() => {
                jest.advanceTimersByTime(300);
            });

            await waitFor(() => {
                expect(screen.getByTestId('user-dropdown')).toBeInTheDocument();
            });

            // user1 should be excluded
            expect(screen.queryByTestId('user-option-user1')).not.toBeInTheDocument();
            // Other users should be present
            expect(screen.getByTestId('user-option-user2')).toBeInTheDocument();
        });

        it('displays selected user after selection', async () => {
            const mockOnChange = jest.fn();
            render(<UserSelector {...defaultProps} onChange={mockOnChange} />);
            const input = screen.getByRole('combobox');

            fireEvent.change(input, {target: {value: 'john'}});

            act(() => {
                jest.advanceTimersByTime(300);
            });

            await waitFor(() => {
                expect(screen.getByTestId('user-dropdown')).toBeInTheDocument();
            });

            fireEvent.click(screen.getByTestId('user-option-user1'));

            // Input should now show selected user's display name
            expect(input).toHaveValue('John Doe');
            // Dropdown should be closed
            expect(screen.queryByTestId('user-dropdown')).not.toBeInTheDocument();
        });
    });

    describe('Task 4: Selection Handling (AC: 2)', () => {
        it('shows clear button when user is selected', async () => {
            render(<UserSelector {...defaultProps} />);
            const input = screen.getByRole('combobox');

            fireEvent.change(input, {target: {value: 'john'}});

            act(() => {
                jest.advanceTimersByTime(300);
            });

            await waitFor(() => {
                expect(screen.getByTestId('user-dropdown')).toBeInTheDocument();
            });

            fireEvent.click(screen.getByTestId('user-option-user1'));

            // Clear button should appear
            expect(screen.getByTestId('clear-button')).toBeInTheDocument();
        });

        it('clears selection when clear button is clicked', async () => {
            const mockOnChange = jest.fn();
            render(<UserSelector {...defaultProps} onChange={mockOnChange} />);
            const input = screen.getByRole('combobox');

            fireEvent.change(input, {target: {value: 'john'}});

            act(() => {
                jest.advanceTimersByTime(300);
            });

            await waitFor(() => {
                expect(screen.getByTestId('user-dropdown')).toBeInTheDocument();
            });

            fireEvent.click(screen.getByTestId('user-option-user1'));

            // Clear the selection
            fireEvent.click(screen.getByTestId('clear-button'));

            // Input should be cleared
            expect(input).toHaveValue('');
            // onChange should be called with empty string
            expect(mockOnChange).toHaveBeenLastCalledWith('');
            // Clear button should disappear
            expect(screen.queryByTestId('clear-button')).not.toBeInTheDocument();
        });

        it('hides clear button when no user is selected', () => {
            render(<UserSelector {...defaultProps} />);
            expect(screen.queryByTestId('clear-button')).not.toBeInTheDocument();
        });
    });

    describe('Task 5: Error State (AC: 4)', () => {
        it('displays error message when error prop is set', () => {
            render(<UserSelector {...defaultProps} error="Please select an approver" />);
            expect(screen.getByText('Please select an approver')).toBeInTheDocument();
        });

        it('applies red border styling when error is present', () => {
            render(<UserSelector {...defaultProps} error="Error message" />);
            const input = screen.getByRole('combobox');
            expect(input).toHaveStyle({
                borderColor: 'var(--error-text, #d24b4e)',
            });
        });

        it('uses --error-text CSS variable for error message', () => {
            render(<UserSelector {...defaultProps} error="Error message" />);
            const errorElement = screen.getByRole('alert');
            expect(errorElement).toHaveStyle({
                color: 'var(--error-text, #d24b4e)',
            });
        });

        it('has aria-invalid set to true when error is present', () => {
            render(<UserSelector {...defaultProps} error="Error message" />);
            const input = screen.getByRole('combobox');
            expect(input).toHaveAttribute('aria-invalid', 'true');
        });

        it('has aria-describedby pointing to error element', () => {
            render(<UserSelector {...defaultProps} error="Error message" />);
            const input = screen.getByRole('combobox');
            expect(input).toHaveAttribute('aria-describedby', 'user-selector-error');
        });

        it('does not show error message when error prop is not set', () => {
            render(<UserSelector {...defaultProps} />);
            expect(screen.queryByRole('alert')).not.toBeInTheDocument();
        });

        it('expects parent to clear error on selection (onChange called)', async () => {
            // AC4: "Error clears when user makes new selection"
            // The parent component should clear error when onChange is called
            const mockOnChange = jest.fn();
            const {rerender} = render(
                <UserSelector {...defaultProps} onChange={mockOnChange} error="Please select" />
            );
            const input = screen.getByRole('combobox');

            // Make a selection
            fireEvent.change(input, {target: {value: 'john'}});
            act(() => {
                jest.advanceTimersByTime(300);
            });

            await waitFor(() => {
                expect(screen.getByTestId('user-dropdown')).toBeInTheDocument();
            });

            fireEvent.click(screen.getByTestId('user-option-user1'));

            // onChange was called - parent should clear error
            expect(mockOnChange).toHaveBeenCalledWith('user1');

            // Simulate parent clearing error after selection
            rerender(<UserSelector {...defaultProps} onChange={mockOnChange} error={undefined} />);

            // Error should now be gone
            expect(screen.queryByRole('alert')).not.toBeInTheDocument();
        });
    });

    describe('Task 4.5: Controlled Component Pattern', () => {
        it('syncs with external value prop when cleared by parent', async () => {
            const mockOnChange = jest.fn();
            // Start with a value to simulate controlled component
            const {rerender} = render(
                <UserSelector {...defaultProps} value="user1" onChange={mockOnChange} />
            );
            const input = screen.getByRole('combobox');

            // Make a selection (simulating user selecting someone)
            fireEvent.change(input, {target: {value: 'john'}});
            act(() => {
                jest.advanceTimersByTime(300);
            });

            await waitFor(() => {
                expect(screen.getByTestId('user-dropdown')).toBeInTheDocument();
            });

            fireEvent.click(screen.getByTestId('user-option-user1'));
            expect(input).toHaveValue('John Doe');
            expect(mockOnChange).toHaveBeenCalledWith('user1');

            // Parent updates to the selected value, then clears it
            rerender(<UserSelector {...defaultProps} value="user1" onChange={mockOnChange} />);
            rerender(<UserSelector {...defaultProps} value="" onChange={mockOnChange} />);

            // Component should sync - input should be cleared
            expect(input).toHaveValue('');
            // Clear button should be gone
            expect(screen.queryByTestId('clear-button')).not.toBeInTheDocument();
        });
    });
});
