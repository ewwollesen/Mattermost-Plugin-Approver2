/**
 * TextArea Component
 * Story 11.3 - Task 2: Create Description TextArea component
 *
 * A reusable text area component with label, character counter, and error states.
 * Uses Mattermost CSS variables for consistent styling.
 *
 * @example
 * <TextArea
 *     value={description}
 *     onChange={(text) => setDescription(text)}
 *     label="What needs approval?"
 *     placeholder="Describe what needs approval..."
 *     maxLength={1000}
 *     error={errors.description}
 * />
 */

import React, {useRef} from 'react';

// Generate unique ID for each TextArea instance (Issue 7 fix)
let textAreaIdCounter = 0;
const generateTextAreaId = () => {
    textAreaIdCounter += 1;
    return `textarea-${textAreaIdCounter}`;
};

/**
 * TextArea component props
 */
export interface TextAreaProps {
    /** Current value */
    value: string;
    /** Callback when value changes */
    onChange: (value: string) => void;
    /** Label text */
    label?: string;
    /** Placeholder text */
    placeholder?: string;
    /** Error message to display */
    error?: string;
    /** Maximum character length */
    maxLength?: number;
    /** Disable the textarea */
    disabled?: boolean;
    /** Number of visible rows (default: 4) */
    rows?: number;
}

/**
 * Inline styles using Mattermost CSS variables
 */
const styles: Record<string, React.CSSProperties> = {
    container: {
        marginBottom: '16px',
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
    textarea: {
        width: '100%',
        padding: '10px 12px',
        borderWidth: '1px',
        borderStyle: 'solid',
        borderColor: 'var(--center-channel-color-16, rgba(61, 60, 64, 0.16))',
        borderRadius: '4px',
        backgroundColor: 'var(--center-channel-bg, #ffffff)',
        color: 'var(--center-channel-color, #3d3c40)',
        fontSize: '14px',
        fontFamily: 'inherit',
        resize: 'vertical',
        outline: 'none',
        transition: 'border-color 0.15s ease',
        boxSizing: 'border-box',
    },
    textareaFocused: {
        borderColor: 'var(--button-bg, #166de0)',
    },
    textareaError: {
        borderColor: 'var(--error-text, #d24b4e)',
    },
    textareaDisabled: {
        backgroundColor: 'var(--center-channel-color-04, rgba(61, 60, 64, 0.04))',
        cursor: 'not-allowed',
    },
    footer: {
        display: 'flex',
        justifyContent: 'space-between',
        alignItems: 'flex-start',
        marginTop: '4px',
    },
    error: {
        color: 'var(--error-text, #d24b4e)',
        fontSize: '12px',
        flex: 1,
    },
    counter: {
        fontSize: '12px',
        color: 'var(--center-channel-color-56, rgba(61, 60, 64, 0.56))',
        marginLeft: '8px',
        whiteSpace: 'nowrap',
    },
    counterWarning: {
        color: 'var(--error-text, #d24b4e)',
    },
};

/**
 * TextArea Component
 */
const TextAreaInner: React.FC<TextAreaProps> = ({
    value,
    onChange,
    label,
    placeholder,
    error,
    maxLength,
    disabled = false,
    rows = 4,
}) => {
    const [isFocused, setIsFocused] = React.useState(false);
    // Issue 7 fix: Generate unique ID per instance for accessibility
    const errorIdRef = useRef<string>(generateTextAreaId() + '-error');

    // Compute textarea styles based on state
    const getTextareaStyle = (): React.CSSProperties => {
        let style = {...styles.textarea};

        if (disabled) {
            style = {...style, ...styles.textareaDisabled};
        }
        if (error) {
            style = {...style, ...styles.textareaError};
        }
        if (isFocused && !error) {
            style = {...style, ...styles.textareaFocused};
        }

        return style;
    };

    // Determine if character count is near limit (> 90%)
    const isNearLimit = maxLength && value.length > maxLength * 0.9;

    // Get counter style
    const getCounterStyle = (): React.CSSProperties => {
        if (isNearLimit) {
            return {...styles.counter, ...styles.counterWarning};
        }
        return styles.counter;
    };

    const handleChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
        onChange(e.target.value);
    };

    return (
        <div style={styles.container} data-testid="textarea-container">
            {label && (
                <label style={styles.label}>
                    {label}
                    <span style={styles.required}>*</span>
                </label>
            )}
            <textarea
                value={value}
                onChange={handleChange}
                onFocus={() => setIsFocused(true)}
                onBlur={() => setIsFocused(false)}
                placeholder={placeholder}
                disabled={disabled}
                rows={rows}
                maxLength={maxLength}
                style={getTextareaStyle()}
                aria-invalid={!!error}
                aria-describedby={error ? errorIdRef.current : undefined}
                data-testid="textarea-input"
            />
            <div style={styles.footer}>
                {error ? (
                    <div id={errorIdRef.current} style={styles.error} role="alert" data-testid="textarea-error">
                        {error}
                    </div>
                ) : (
                    <div />
                )}
                {maxLength && (
                    <div style={getCounterStyle()} data-testid="character-counter">
                        {value.length}/{maxLength}
                    </div>
                )}
            </div>
        </div>
    );
};

// Memoize for performance
const TextArea = React.memo(TextAreaInner);

export default TextArea;
