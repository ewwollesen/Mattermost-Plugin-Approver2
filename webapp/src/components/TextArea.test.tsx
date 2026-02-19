/**
 * TextArea Component Tests
 * Story 11.3 - Task 2: Create Description TextArea component
 */

import React from 'react';
import {render, screen, fireEvent} from '@testing-library/react';
import TextArea from './TextArea';

describe('TextArea Component', () => {
    const defaultProps = {
        value: '',
        onChange: jest.fn(),
    };

    beforeEach(() => {
        jest.clearAllMocks();
    });

    describe('Basic Rendering', () => {
        it('renders textarea input', () => {
            render(<TextArea {...defaultProps} />);
            expect(screen.getByTestId('textarea-input')).toBeInTheDocument();
        });

        it('renders with label when provided', () => {
            render(<TextArea {...defaultProps} label="Description" />);
            expect(screen.getByText('Description')).toBeInTheDocument();
        });

        it('renders required indicator when label provided', () => {
            render(<TextArea {...defaultProps} label="Description" />);
            expect(screen.getByText('*')).toBeInTheDocument();
        });

        it('renders with placeholder when provided', () => {
            render(<TextArea {...defaultProps} placeholder="Enter text..." />);
            expect(screen.getByPlaceholderText('Enter text...')).toBeInTheDocument();
        });

        it('renders with custom rows', () => {
            render(<TextArea {...defaultProps} rows={6} />);
            const textarea = screen.getByTestId('textarea-input');
            expect(textarea).toHaveAttribute('rows', '6');
        });

        it('uses default rows of 4', () => {
            render(<TextArea {...defaultProps} />);
            const textarea = screen.getByTestId('textarea-input');
            expect(textarea).toHaveAttribute('rows', '4');
        });
    });

    describe('Character Counter', () => {
        it('shows character counter when maxLength provided', () => {
            render(<TextArea {...defaultProps} maxLength={1000} />);
            expect(screen.getByTestId('character-counter')).toBeInTheDocument();
            expect(screen.getByText('0/1000')).toBeInTheDocument();
        });

        it('does not show character counter when maxLength not provided', () => {
            render(<TextArea {...defaultProps} />);
            expect(screen.queryByTestId('character-counter')).not.toBeInTheDocument();
        });

        it('updates counter as user types', () => {
            render(<TextArea {...defaultProps} value="Hello" maxLength={1000} />);
            expect(screen.getByText('5/1000')).toBeInTheDocument();
        });

        it('shows warning style when near limit (>90%)', () => {
            render(<TextArea {...defaultProps} value={'a'.repeat(950)} maxLength={1000} />);
            const counter = screen.getByTestId('character-counter');
            expect(counter).toHaveStyle({color: 'var(--error-text, #d24b4e)'});
        });

        it('does not show warning style when below 90%', () => {
            render(<TextArea {...defaultProps} value={'a'.repeat(500)} maxLength={1000} />);
            const counter = screen.getByTestId('character-counter');
            expect(counter).toHaveStyle({color: 'var(--center-channel-color-56, rgba(61, 60, 64, 0.56))'});
        });
    });

    describe('Error State', () => {
        it('displays error message when error prop set', () => {
            render(<TextArea {...defaultProps} error="This field is required" />);
            expect(screen.getByText('This field is required')).toBeInTheDocument();
        });

        it('applies red border when error present', () => {
            render(<TextArea {...defaultProps} error="Error" />);
            const textarea = screen.getByTestId('textarea-input');
            expect(textarea).toHaveStyle({
                borderColor: 'var(--error-text, #d24b4e)',
            });
        });

        it('has aria-invalid set to true when error present', () => {
            render(<TextArea {...defaultProps} error="Error" />);
            const textarea = screen.getByTestId('textarea-input');
            expect(textarea).toHaveAttribute('aria-invalid', 'true');
        });

        it('has aria-describedby pointing to error element', () => {
            render(<TextArea {...defaultProps} error="Error" />);
            const textarea = screen.getByTestId('textarea-input');
            // Issue 7 fix: ID is now dynamically generated, just verify it exists and points to error
            const describedBy = textarea.getAttribute('aria-describedby');
            expect(describedBy).toBeTruthy();
            expect(describedBy).toMatch(/^textarea-\d+-error$/);
        });

        it('error message has role="alert"', () => {
            render(<TextArea {...defaultProps} error="Error" />);
            expect(screen.getByRole('alert')).toBeInTheDocument();
        });

        it('does not show error message when error prop not set', () => {
            render(<TextArea {...defaultProps} />);
            expect(screen.queryByRole('alert')).not.toBeInTheDocument();
        });
    });

    describe('Disabled State', () => {
        it('disables textarea when disabled prop is true', () => {
            render(<TextArea {...defaultProps} disabled={true} />);
            const textarea = screen.getByTestId('textarea-input');
            expect(textarea).toBeDisabled();
        });

        it('applies disabled styling', () => {
            render(<TextArea {...defaultProps} disabled={true} />);
            const textarea = screen.getByTestId('textarea-input');
            expect(textarea).toHaveStyle({
                backgroundColor: 'var(--center-channel-color-04, rgba(61, 60, 64, 0.04))',
                cursor: 'not-allowed',
            });
        });
    });

    describe('User Interaction', () => {
        it('calls onChange when user types', () => {
            const mockOnChange = jest.fn();
            render(<TextArea {...defaultProps} onChange={mockOnChange} />);
            const textarea = screen.getByTestId('textarea-input');

            fireEvent.change(textarea, {target: {value: 'Hello'}});

            expect(mockOnChange).toHaveBeenCalledWith('Hello');
        });

        it('displays provided value', () => {
            render(<TextArea {...defaultProps} value="Test value" />);
            const textarea = screen.getByTestId('textarea-input');
            expect(textarea).toHaveValue('Test value');
        });
    });

    describe('Focus State', () => {
        it('applies focus styling when focused', () => {
            render(<TextArea {...defaultProps} />);
            const textarea = screen.getByTestId('textarea-input');

            fireEvent.focus(textarea);

            expect(textarea).toHaveStyle({
                borderColor: 'var(--button-bg, #166de0)',
            });
        });

        it('removes focus styling when blurred', () => {
            render(<TextArea {...defaultProps} />);
            const textarea = screen.getByTestId('textarea-input');

            fireEvent.focus(textarea);
            fireEvent.blur(textarea);

            expect(textarea).toHaveStyle({
                borderColor: 'var(--center-channel-color-16, rgba(61, 60, 64, 0.16))',
            });
        });

        it('error style takes precedence over focus style', () => {
            render(<TextArea {...defaultProps} error="Error" />);
            const textarea = screen.getByTestId('textarea-input');

            fireEvent.focus(textarea);

            expect(textarea).toHaveStyle({
                borderColor: 'var(--error-text, #d24b4e)',
            });
        });
    });

    describe('Mattermost Styling', () => {
        it('uses Mattermost CSS variables', () => {
            render(<TextArea {...defaultProps} />);
            const textarea = screen.getByTestId('textarea-input');
            expect(textarea).toHaveStyle({
                backgroundColor: 'var(--center-channel-bg, #ffffff)',
                color: 'var(--center-channel-color, #3d3c40)',
            });
        });
    });
});
