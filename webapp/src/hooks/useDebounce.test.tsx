/**
 * useDebounce Hook Tests
 * Story 11.2: User Selector Component - Task 2.3
 */

import React, {useState} from 'react';
import {render, screen, fireEvent, act} from '@testing-library/react';
import {useDebounce} from './useDebounce';

// Test component that uses the hook
interface TestComponentProps {
    initialValue: string;
    delay?: number;
}

const TestComponent: React.FC<TestComponentProps> = ({initialValue, delay = 300}) => {
    const [value, setValue] = useState(initialValue);
    const debouncedValue = useDebounce(value, delay);

    return (
        <div>
            <input
                data-testid="input"
                value={value}
                onChange={(e) => setValue(e.target.value)}
            />
            <span data-testid="debounced">{debouncedValue}</span>
            <span data-testid="current">{value}</span>
        </div>
    );
};

describe('useDebounce Hook', () => {
    beforeEach(() => {
        jest.useFakeTimers();
    });

    afterEach(() => {
        jest.useRealTimers();
    });

    it('returns initial value immediately', () => {
        render(<TestComponent initialValue="initial" />);

        expect(screen.getByTestId('debounced').textContent).toBe('initial');
    });

    it('does not update debounced value before delay', () => {
        render(<TestComponent initialValue="initial" delay={300} />);

        const input = screen.getByTestId('input');

        fireEvent.change(input, {target: {value: 'updated'}});

        // Current value updates immediately
        expect(screen.getByTestId('current').textContent).toBe('updated');
        // Debounced value stays the same
        expect(screen.getByTestId('debounced').textContent).toBe('initial');
    });

    it('updates debounced value after delay', () => {
        render(<TestComponent initialValue="initial" delay={300} />);

        const input = screen.getByTestId('input');

        fireEvent.change(input, {target: {value: 'updated'}});

        act(() => {
            jest.advanceTimersByTime(300);
        });

        expect(screen.getByTestId('debounced').textContent).toBe('updated');
    });

    it('resets timer on rapid changes', () => {
        render(<TestComponent initialValue="initial" delay={300} />);

        const input = screen.getByTestId('input');

        // First change
        fireEvent.change(input, {target: {value: 'a'}});

        act(() => {
            jest.advanceTimersByTime(100);
        });

        // Second change before delay completes
        fireEvent.change(input, {target: {value: 'ab'}});

        act(() => {
            jest.advanceTimersByTime(100);
        });

        // Third change
        fireEvent.change(input, {target: {value: 'abc'}});

        // Debounced should still be initial
        expect(screen.getByTestId('debounced').textContent).toBe('initial');

        // Complete the delay
        act(() => {
            jest.advanceTimersByTime(300);
        });

        // Should now be final value
        expect(screen.getByTestId('debounced').textContent).toBe('abc');
    });
});
