import React, {useEffect, useRef, useCallback, useState} from 'react';

/**
 * Modal Component Props
 * Story 11.1 - AC2: Modal Container Component
 */
export interface ModalProps {
    /** Whether the modal is visible */
    visible: boolean;
    /** Callback when modal should close */
    onClose: () => void;
    /** Modal title displayed in header */
    title: string;
    /** Modal content */
    children: React.ReactNode;
    /** Custom width (default: 480px) */
    width?: string;
}

/**
 * Modal Component
 *
 * A reusable modal component that:
 * - Handles overlay click to close (AC2)
 * - Handles Escape key to close (AC2)
 * - Implements focus trap (AC2)
 * - Uses Mattermost CSS variables for styling (AC2)
 *
 * @example
 * <Modal visible={true} onClose={() => setOpen(false)} title="My Modal">
 *   <p>Modal content here</p>
 * </Modal>
 */
// Generate unique ID for each modal instance (Issue 4 fix)
let modalIdCounter = 0;
const generateModalId = () => {
    modalIdCounter += 1;
    return `modal-${modalIdCounter}`;
};

// Selector for focusable elements within the modal
const FOCUSABLE_SELECTOR = [
    'button:not([disabled])',
    'a[href]',
    'input:not([disabled])',
    'select:not([disabled])',
    'textarea:not([disabled])',
    '[tabindex]:not([tabindex="-1"])',
].join(', ');

const Modal: React.FC<ModalProps> = ({
    visible,
    onClose,
    title,
    children,
    width = '480px',
}) => {
    const modalRef = useRef<HTMLDivElement>(null);
    const previousActiveElement = useRef<Element | null>(null);
    const modalIdRef = useRef<string>(generateModalId());

    // Issue 5 fix: Track hover/focus state for close button
    const [isCloseButtonHovered, setIsCloseButtonHovered] = useState(false);
    const [isCloseButtonFocused, setIsCloseButtonFocused] = useState(false);

    // Handle Escape key to close modal (AC2)
    const handleKeyDown = useCallback((event: KeyboardEvent) => {
        if (event.key === 'Escape') {
            onClose();
        }
    }, [onClose]);

    // Issue 1 fix: Implement proper focus trap
    const handleFocusTrap = useCallback((event: KeyboardEvent) => {
        if (event.key !== 'Tab' || !modalRef.current) {
            return;
        }

        const focusableElements = modalRef.current.querySelectorAll<HTMLElement>(FOCUSABLE_SELECTOR);
        if (focusableElements.length === 0) {
            // No focusable elements, prevent tab from leaving modal
            event.preventDefault();
            return;
        }

        const firstElement = focusableElements[0];
        const lastElement = focusableElements[focusableElements.length - 1];

        if (event.shiftKey) {
            // Shift+Tab: If on first element, wrap to last
            if (document.activeElement === firstElement) {
                event.preventDefault();
                lastElement.focus();
            }
        } else {
            // Tab: If on last element, wrap to first
            if (document.activeElement === lastElement) {
                event.preventDefault();
                firstElement.focus();
            }
        }
    }, []);

    // Handle overlay click to close (AC2)
    const handleOverlayClick = useCallback((event: React.MouseEvent<HTMLDivElement>) => {
        // Only close if clicking the overlay itself, not the modal content
        if (event.target === event.currentTarget) {
            onClose();
        }
    }, [onClose]);

    // Stop propagation when clicking inside modal content
    const handleContentClick = useCallback((event: React.MouseEvent<HTMLDivElement>) => {
        event.stopPropagation();
    }, []);

    // Set up and clean up event listeners (AC2)
    useEffect(() => {
        if (visible) {
            // Store the currently focused element to restore later
            previousActiveElement.current = document.activeElement;

            // Add escape key listener
            document.addEventListener('keydown', handleKeyDown);

            // Issue 1 fix: Add focus trap listener
            document.addEventListener('keydown', handleFocusTrap);

            // Focus the modal container
            if (modalRef.current) {
                modalRef.current.focus();
            }

            return () => {
                document.removeEventListener('keydown', handleKeyDown);
                document.removeEventListener('keydown', handleFocusTrap);

                // Restore focus to previously focused element
                if (previousActiveElement.current instanceof HTMLElement) {
                    previousActiveElement.current.focus();
                }
            };
        }
        return undefined;
    }, [visible, handleKeyDown, handleFocusTrap]);

    // Don't render if not visible
    if (!visible) {
        return null;
    }

    // Issue 4 fix: Use unique ID for this modal instance
    const titleId = `${modalIdRef.current}-title`;

    // Issue 5 fix: Dynamic close button styles based on hover/focus state
    const closeButtonStyle: React.CSSProperties = {
        ...styles.closeButton,
        ...(isCloseButtonHovered && styles.closeButtonHover),
        ...(isCloseButtonFocused && styles.closeButtonFocus),
    };

    return (
        <div
            data-testid="modal-overlay"
            onClick={handleOverlayClick}
            style={styles.overlay}
        >
            <div
                ref={modalRef}
                data-testid="modal-content"
                onClick={handleContentClick}
                role="dialog"
                aria-modal="true"
                aria-labelledby={titleId}
                tabIndex={-1}
                style={{...styles.modal, width}}
            >
                <div style={styles.header}>
                    <h2 id={titleId} style={styles.title}>
                        {title}
                    </h2>
                    <button
                        type="button"
                        onClick={onClose}
                        aria-label="Close modal"
                        style={closeButtonStyle}
                        onMouseEnter={() => setIsCloseButtonHovered(true)}
                        onMouseLeave={() => setIsCloseButtonHovered(false)}
                        onFocus={() => setIsCloseButtonFocused(true)}
                        onBlur={() => setIsCloseButtonFocused(false)}
                    >
                        &times;
                    </button>
                </div>
                <div style={styles.body}>
                    {children}
                </div>
            </div>
        </div>
    );
};

/**
 * Inline styles using Mattermost CSS variables where possible
 * These provide the base styling; components can override as needed
 */
const styles: Record<string, React.CSSProperties> = {
    overlay: {
        position: 'fixed',
        top: 0,
        left: 0,
        right: 0,
        bottom: 0,
        backgroundColor: 'rgba(0, 0, 0, 0.5)',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        zIndex: 10000,
    },
    modal: {
        backgroundColor: 'var(--center-channel-bg, #ffffff)',
        color: 'var(--center-channel-color, #3d3c40)',
        borderRadius: '8px',
        boxShadow: '0 4px 24px rgba(0, 0, 0, 0.2)',
        maxHeight: '90vh',
        overflow: 'auto',
        outline: 'none', // Remove focus outline, we handle focus visually
    },
    header: {
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'space-between',
        padding: '16px 20px',
        borderBottom: '1px solid var(--center-channel-color-08, rgba(61, 60, 64, 0.08))',
    },
    title: {
        margin: 0,
        fontSize: '18px',
        fontWeight: 600,
        color: 'var(--center-channel-color, #3d3c40)',
    },
    closeButton: {
        backgroundColor: 'transparent',
        border: 'none',
        fontSize: '24px',
        cursor: 'pointer',
        padding: '4px 8px',
        color: 'var(--center-channel-color-56, rgba(61, 60, 64, 0.56))',
        borderRadius: '4px',
        lineHeight: 1,
        transition: 'background-color 0.15s ease, color 0.15s ease',
        outline: 'none',
    },
    // Issue 5 fix: Hover state for close button
    closeButtonHover: {
        backgroundColor: 'var(--center-channel-color-08, rgba(61, 60, 64, 0.08))',
        color: 'var(--center-channel-color, #3d3c40)',
    },
    // Issue 5 fix: Focus state for close button
    closeButtonFocus: {
        backgroundColor: 'var(--center-channel-color-08, rgba(61, 60, 64, 0.08))',
        color: 'var(--center-channel-color, #3d3c40)',
        outline: '2px solid var(--button-bg, #166de0)',
        outlineOffset: '1px',
    },
    body: {
        padding: '20px',
    },
};

export default Modal;
