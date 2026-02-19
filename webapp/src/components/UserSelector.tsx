/**
 * UserSelector Component
 * Story 11.2: User Selector Component
 *
 * A searchable user selector with autocomplete functionality.
 * Uses Mattermost's user autocomplete API for search.
 *
 * @example
 * <UserSelector
 *     value={selectedUserId}
 *     onChange={(userId) => setSelectedUserId(userId)}
 *     label="Select Approver"
 *     placeholder="Search users..."
 *     error={errors.approver}
 *     excludeUserIds={[currentUserId]}
 * />
 */

import React, {useState, useEffect, useRef, useCallback} from 'react';
import {searchUsers, UserOption} from '../api/users';
import {useDebounce} from '../hooks/useDebounce';

export type {UserOption};

/**
 * UserSelector component props
 */
export interface UserSelectorProps {
    /** Currently selected user ID */
    value: string;
    /** Callback when selection changes */
    onChange: (userId: string) => void;
    /** Error message to display */
    error?: string;
    /** Label text */
    label?: string;
    /** Placeholder text for search input */
    placeholder?: string;
    /** Disable the selector */
    disabled?: boolean;
    /** User IDs to exclude from results (e.g., current user for self-approval prevention) */
    excludeUserIds?: string[];
}

/**
 * Inline styles using Mattermost CSS variables
 */
const styles: Record<string, React.CSSProperties> = {
    container: {
        position: 'relative',
        width: '100%',
    },
    label: {
        display: 'block',
        marginBottom: '8px',
        fontWeight: 600,
        fontSize: '14px',
        color: 'var(--center-channel-color, #3d3c40)',
    },
    required: {
        color: 'var(--error-text, #d24b4e)',
        marginLeft: '4px',
    },
    inputContainer: {
        position: 'relative',
        display: 'flex',
        alignItems: 'center',
    },
    searchIcon: {
        position: 'absolute',
        left: '12px',
        color: 'var(--center-channel-color-56, rgba(61, 60, 64, 0.56))',
        pointerEvents: 'none',
        fontSize: '16px',
    },
    input: {
        width: '100%',
        padding: '10px 12px',
        paddingLeft: '36px',
        borderWidth: '1px',
        borderStyle: 'solid',
        borderColor: 'var(--center-channel-color-16, rgba(61, 60, 64, 0.16))',
        borderRadius: '4px',
        backgroundColor: 'var(--center-channel-bg, #ffffff)',
        color: 'var(--center-channel-color, #3d3c40)',
        fontSize: '14px',
        outline: 'none',
        transition: 'border-color 0.15s ease',
    },
    inputFocused: {
        borderColor: 'var(--button-bg, #166de0)',
    },
    inputError: {
        borderColor: 'var(--error-text, #d24b4e)',
    },
    inputDisabled: {
        backgroundColor: 'var(--center-channel-color-04, rgba(61, 60, 64, 0.04))',
        cursor: 'not-allowed',
    },
    error: {
        color: 'var(--error-text, #d24b4e)',
        fontSize: '12px',
        marginTop: '4px',
    },
    dropdown: {
        position: 'absolute',
        top: '100%',
        left: 0,
        right: 0,
        marginTop: '4px',
        backgroundColor: 'var(--center-channel-bg, #ffffff)',
        border: '1px solid var(--center-channel-color-16, rgba(61, 60, 64, 0.16))',
        borderRadius: '4px',
        boxShadow: '0 4px 6px rgba(0, 0, 0, 0.1)',
        zIndex: 1000,
        maxHeight: '200px',
        overflowY: 'auto',
    },
    userOption: {
        display: 'flex',
        alignItems: 'center',
        padding: '8px 12px',
        cursor: 'pointer',
        transition: 'background-color 0.1s ease',
    },
    userOptionHighlighted: {
        backgroundColor: 'var(--button-bg-08, rgba(22, 109, 224, 0.08))',
    },
    avatar: {
        width: '24px',
        height: '24px',
        borderRadius: '50%',
        marginRight: '8px',
        backgroundColor: 'var(--center-channel-color-16, rgba(61, 60, 64, 0.16))',
    },
    userInfo: {
        display: 'flex',
        flexDirection: 'column',
    },
    displayName: {
        fontSize: '14px',
        fontWeight: 500,
        color: 'var(--center-channel-color, #3d3c40)',
    },
    username: {
        fontSize: '12px',
        color: 'var(--center-channel-color-56, rgba(61, 60, 64, 0.56))',
    },
    loadingSpinner: {
        display: 'flex',
        justifyContent: 'center',
        alignItems: 'center',
        padding: '16px',
        color: 'var(--center-channel-color-56, rgba(61, 60, 64, 0.56))',
    },
    noResults: {
        padding: '16px',
        textAlign: 'center',
        color: 'var(--center-channel-color-56, rgba(61, 60, 64, 0.56))',
        fontSize: '14px',
    },
    clearButton: {
        position: 'absolute',
        right: '12px',
        background: 'none',
        border: 'none',
        cursor: 'pointer',
        padding: '4px',
        color: 'var(--center-channel-color-56, rgba(61, 60, 64, 0.56))',
        fontSize: '14px',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
    },
    avatarFallback: {
        width: '24px',
        height: '24px',
        borderRadius: '50%',
        marginRight: '8px',
        backgroundColor: 'var(--center-channel-color-16, rgba(61, 60, 64, 0.16))',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        fontSize: '12px',
        color: 'var(--center-channel-color-56, rgba(61, 60, 64, 0.56))',
    },
};

/**
 * UserSelector Component
 * Task 1: Create UserSelector component structure (AC: 1, 3)
 * Task 3: Build dropdown results list (AC: 1, 3)
 * Task 4.5: Controlled component pattern - syncs with external value prop
 */
const UserSelectorInner: React.FC<UserSelectorProps> = ({
    value,
    onChange,
    error,
    label,
    placeholder = 'Search users...',
    disabled = false,
    excludeUserIds = [],
}) => {
    const [searchTerm, setSearchTerm] = useState('');
    const [isFocused, setIsFocused] = useState(false);
    const [isDropdownOpen, setIsDropdownOpen] = useState(false);
    const [users, setUsers] = useState<UserOption[]>([]);
    const [isLoading, setIsLoading] = useState(false);
    const [highlightedIndex, setHighlightedIndex] = useState(-1);
    const [selectedUser, setSelectedUser] = useState<UserOption | null>(null);

    const inputRef = useRef<HTMLInputElement>(null);
    const containerRef = useRef<HTMLDivElement>(null);
    const excludeUserIdsRef = useRef<string[]>(excludeUserIds);

    // Keep ref in sync with prop
    useEffect(() => {
        excludeUserIdsRef.current = excludeUserIds;
    }, [excludeUserIds]);

    // Track previous value for controlled component sync (Task 4.5)
    const prevValueRef = useRef(value);

    // Sync with controlled value prop (Task 4.5)
    // When value changes from non-empty to empty externally, reset internal state
    useEffect(() => {
        const prevValue = prevValueRef.current;
        prevValueRef.current = value;

        // Only reset if value changed from something to empty (external clear)
        if (prevValue !== '' && value === '' && selectedUser !== null) {
            setSelectedUser(null);
            setSearchTerm('');
        }
    }, [value, selectedUser]);

    // Debounce the search term
    const debouncedSearchTerm = useDebounce(searchTerm, 300);

    // Fetch users when debounced search term changes
    useEffect(() => {
        if (debouncedSearchTerm.length < 2) {
            setUsers([]);
            setIsDropdownOpen(false);
            return;
        }

        let cancelled = false;

        const fetchUsers = async () => {
            setIsLoading(true);
            setIsDropdownOpen(true);

            try {
                const results = await searchUsers(debouncedSearchTerm);
                if (cancelled) return;

                // Filter out excluded users
                const filteredResults = results.filter(
                    (user) => !excludeUserIdsRef.current.includes(user.id)
                );
                setUsers(filteredResults);
            } catch (err) {
                if (cancelled) return;
                console.error('Error fetching users:', err);
                setUsers([]);
            } finally {
                if (!cancelled) {
                    setIsLoading(false);
                }
            }
        };

        fetchUsers();

        return () => {
            cancelled = true;
        };
    }, [debouncedSearchTerm]);

    // Handle click outside to close dropdown
    useEffect(() => {
        const handleClickOutside = (event: MouseEvent) => {
            if (
                containerRef.current &&
                !containerRef.current.contains(event.target as Node)
            ) {
                setIsDropdownOpen(false);
            }
        };

        document.addEventListener('mousedown', handleClickOutside);
        return () => {
            document.removeEventListener('mousedown', handleClickOutside);
        };
    }, []);

    // Compute input styles based on state
    const getInputStyle = (): React.CSSProperties => {
        let style = {...styles.input};

        if (disabled) {
            style = {...style, ...styles.inputDisabled};
        }
        if (error) {
            style = {...style, ...styles.inputError};
        }
        if (isFocused && !error) {
            style = {...style, ...styles.inputFocused};
        }

        return style;
    };

    const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        setSearchTerm(e.target.value);
        setSelectedUser(null);
        setHighlightedIndex(-1);
    };

    const handleFocus = () => {
        setIsFocused(true);
    };

    const handleBlur = () => {
        setIsFocused(false);
    };

    const handleSelectUser = useCallback((user: UserOption) => {
        setSelectedUser(user);
        setSearchTerm(user.displayName);
        setIsDropdownOpen(false);
        onChange(user.id);
    }, [onChange]);

    const handleClear = useCallback(() => {
        setSelectedUser(null);
        setSearchTerm('');
        setHighlightedIndex(-1);
        onChange('');
        inputRef.current?.focus();
    }, [onChange]);

    const handleKeyDown = useCallback((e: React.KeyboardEvent<HTMLInputElement>) => {
        if (!isDropdownOpen) return;

        switch (e.key) {
            case 'Escape':
                setIsDropdownOpen(false);
                setHighlightedIndex(-1);
                break;
            case 'ArrowDown':
                e.preventDefault();
                setHighlightedIndex((prev) =>
                    prev < users.length - 1 ? prev + 1 : prev
                );
                break;
            case 'ArrowUp':
                e.preventDefault();
                setHighlightedIndex((prev) => (prev > 0 ? prev - 1 : prev));
                break;
            case 'Enter':
                e.preventDefault();
                if (highlightedIndex >= 0 && highlightedIndex < users.length) {
                    handleSelectUser(users[highlightedIndex]);
                }
                break;
        }
    }, [isDropdownOpen, users, highlightedIndex, handleSelectUser]);

    const getOptionStyle = (index: number): React.CSSProperties => {
        return {
            ...styles.userOption,
            ...(index === highlightedIndex ? styles.userOptionHighlighted : {}),
        };
    };

    return (
        <div data-testid="user-selector" style={styles.container} ref={containerRef}>
            {label && (
                <label style={styles.label}>
                    {label}
                    <span style={styles.required}>*</span>
                </label>
            )}
            <div style={styles.inputContainer}>
                <span data-testid="search-icon" style={styles.searchIcon}>
                    &#128269;
                </span>
                <input
                    ref={inputRef}
                    type="text"
                    value={searchTerm}
                    onChange={handleInputChange}
                    onFocus={handleFocus}
                    onBlur={handleBlur}
                    onKeyDown={handleKeyDown}
                    placeholder={placeholder}
                    disabled={disabled}
                    style={getInputStyle()}
                    aria-label={label || 'Search users'}
                    aria-invalid={!!error}
                    aria-describedby={error ? 'user-selector-error' : undefined}
                    aria-expanded={isDropdownOpen}
                    aria-autocomplete="list"
                    aria-activedescendant={highlightedIndex >= 0 && users[highlightedIndex] ? `user-option-${users[highlightedIndex].id}` : undefined}
                    role="combobox"
                />
                {selectedUser && !disabled && (
                    <button
                        data-testid="clear-button"
                        type="button"
                        onClick={handleClear}
                        style={styles.clearButton}
                        aria-label="Clear selection"
                    >
                        &#10005;
                    </button>
                )}
            </div>

            {/* Dropdown */}
            {isDropdownOpen && (
                <div data-testid="user-dropdown" style={styles.dropdown} role="listbox">
                    {isLoading ? (
                        <div data-testid="loading-spinner" style={styles.loadingSpinner}>
                            Loading...
                        </div>
                    ) : users.length === 0 ? (
                        <div style={styles.noResults}>No users found</div>
                    ) : (
                        users.map((user, index) => (
                            <div
                                key={user.id}
                                id={`user-option-${user.id}`}
                                data-testid={`user-option-${user.id}`}
                                data-highlighted={index === highlightedIndex ? 'true' : 'false'}
                                style={getOptionStyle(index)}
                                onClick={() => handleSelectUser(user)}
                                onMouseEnter={() => setHighlightedIndex(index)}
                                role="option"
                                aria-selected={index === highlightedIndex}
                            >
                                <img
                                    data-testid={`user-avatar-${user.id}`}
                                    src={user.avatarUrl}
                                    alt={`${user.displayName}'s avatar`}
                                    style={styles.avatar}
                                    onError={(e) => {
                                        // Hide broken image and show initials fallback
                                        (e.target as HTMLImageElement).style.display = 'none';
                                    }}
                                />
                                <div style={styles.userInfo}>
                                    <span style={styles.displayName}>{user.displayName}</span>
                                    <span style={styles.username}>@{user.username}</span>
                                </div>
                            </div>
                        ))
                    )}
                </div>
            )}

            {error && (
                <div id="user-selector-error" style={styles.error} role="alert">
                    {error}
                </div>
            )}
        </div>
    );
};

// Memoize for performance in forms (prevents unnecessary re-renders)
const UserSelector = React.memo(UserSelectorInner);

export default UserSelector;
