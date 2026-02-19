/**
 * useDebounce Hook
 * Story 11.2: User Selector Component - Task 2.3
 *
 * Debounces a value with a configurable delay.
 */

import {useState, useEffect} from 'react';

/**
 * Custom hook to debounce a value
 *
 * @param value - The value to debounce
 * @param delay - Delay in milliseconds (default: 300ms)
 * @returns The debounced value
 *
 * @example
 * const debouncedSearch = useDebounce(searchTerm, 300);
 * useEffect(() => {
 *   // This runs 300ms after searchTerm stops changing
 *   fetchResults(debouncedSearch);
 * }, [debouncedSearch]);
 */
export const useDebounce = <T>(value: T, delay = 300): T => {
    const [debouncedValue, setDebouncedValue] = useState<T>(value);

    useEffect(() => {
        const timer = setTimeout(() => {
            setDebouncedValue(value);
        }, delay);

        return () => {
            clearTimeout(timer);
        };
    }, [value, delay]);

    return debouncedValue;
};

export default useDebounce;
