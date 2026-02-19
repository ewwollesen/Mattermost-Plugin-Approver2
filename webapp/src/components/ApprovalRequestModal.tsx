/**
 * ApprovalRequestModal Component
 * Story 11.3: Approval Request Modal Component
 *
 * A modal for creating approval requests with:
 * - User selector for choosing an approver (excludes current user)
 * - Description text area with character limit
 * - Client-side validation
 * - Submit and Cancel buttons
 *
 * @example
 * <ApprovalRequestModal
 *     visible={true}
 *     onClose={() => setModalOpen(false)}
 *     channelId="channel123"
 *     teamId="team456"
 *     currentUserId="user789"
 * />
 */

import React, {useState, useCallback, useRef, useEffect} from 'react';
import Modal from './Modal';
import UserSelector from './UserSelector';
import TextArea from './TextArea';

/**
 * Validation error messages (match native dialog)
 */
const VALIDATION_MESSAGES = {
    approverRequired: 'Please select an approver',
    descriptionRequired: 'Please describe what needs approval',
    descriptionTooLong: 'Description must be 1000 characters or less',
};

/**
 * ApprovalRequestModal props
 */
export interface ApprovalRequestModalProps {
    /** Whether the modal is visible */
    visible: boolean;
    /** Callback when modal should close */
    onClose: () => void;
    /** ID of the channel where approval will be created */
    channelId: string;
    /** ID of the team */
    teamId: string;
    /** Current user's ID (excluded from approver selection) */
    currentUserId: string;
}

/**
 * Form state interface
 */
interface FormState {
    approverId: string;
    description: string;
    errors: {
        approver?: string;
        description?: string;
    };
    submitting: boolean;
}

/**
 * Initial form state
 */
const initialFormState: FormState = {
    approverId: '',
    description: '',
    errors: {},
    submitting: false,
};

/**
 * Button styles following Mattermost patterns
 */
const buttonStyles: Record<string, React.CSSProperties> = {
    base: {
        padding: '10px 16px',
        borderRadius: '4px',
        fontSize: '14px',
        fontWeight: 600,
        cursor: 'pointer',
        border: 'none',
        transition: 'background-color 0.15s ease, opacity 0.15s ease',
        minWidth: '80px',
    },
    primary: {
        backgroundColor: 'var(--button-bg, #166de0)',
        color: 'var(--button-color, #ffffff)',
    },
    secondary: {
        backgroundColor: 'transparent',
        color: 'var(--center-channel-color, #3d3c40)',
        border: '1px solid var(--center-channel-color-16, rgba(61, 60, 64, 0.16))',
    },
    disabled: {
        opacity: 0.5,
        cursor: 'not-allowed',
    },
};

/**
 * Component styles
 */
const styles: Record<string, React.CSSProperties> = {
    form: {
        display: 'flex',
        flexDirection: 'column',
        gap: '16px',
    },
    buttonRow: {
        display: 'flex',
        justifyContent: 'flex-end',
        gap: '12px',
        marginTop: '8px',
        paddingTop: '16px',
        borderTop: '1px solid var(--center-channel-color-08, rgba(61, 60, 64, 0.08))',
    },
};

/**
 * ApprovalRequestModal Component
 */
const ApprovalRequestModalInner: React.FC<ApprovalRequestModalProps> = ({
    visible,
    onClose,
    channelId,
    teamId,
    currentUserId,
}) => {
    const [form, setForm] = useState<FormState>(initialFormState);

    // Track mounted state to prevent memory leak on async operations (Issue 6 fix)
    const isMountedRef = useRef(true);

    useEffect(() => {
        isMountedRef.current = true;
        return () => {
            isMountedRef.current = false;
        };
    }, []);

    /**
     * Reset form when modal closes or opens
     */
    useEffect(() => {
        if (!visible) {
            // Reset form state when modal closes
            setForm(initialFormState);
        }
    }, [visible]);

    /**
     * Task 4: Validate form fields
     * Returns true if valid, false otherwise
     */
    const validate = useCallback((): boolean => {
        const errors: FormState['errors'] = {};

        // Validate approver
        if (!form.approverId) {
            errors.approver = VALIDATION_MESSAGES.approverRequired;
        }

        // Validate description
        if (!form.description.trim()) {
            errors.description = VALIDATION_MESSAGES.descriptionRequired;
        } else if (form.description.length > 1000) {
            errors.description = VALIDATION_MESSAGES.descriptionTooLong;
        }

        setForm(f => ({...f, errors}));
        return Object.keys(errors).length === 0;
    }, [form.approverId, form.description]);

    /**
     * Task 5: Handle form submission
     * Story 11.4: Call API endpoint to create approval request
     * Issue 6 fix: Check isMountedRef before state updates after async operations
     */
    const handleSubmit = useCallback(async () => {
        if (!validate()) {
            return;
        }

        setForm(f => ({...f, submitting: true}));

        try {
            // Story 11.4: Call API endpoint to create approval
            const response = await fetch('/plugins/com.mattermost.plugin-approver2/api/v1/approval/new', {
                method: 'POST',
                headers: {'Content-Type': 'application/json'},
                body: JSON.stringify({
                    channel_id: channelId,
                    team_id: teamId,
                    approver_id: form.approverId,
                    description: form.description,
                }),
            });

            // Check if component is still mounted (Issue 6 fix: prevent memory leak)
            if (!isMountedRef.current) {
                return;
            }

            const data = await response.json();

            if (data.success) {
                // Success - close the modal
                onClose();
            } else if (data.errors) {
                // Field-specific validation errors from server
                setForm(f => ({
                    ...f,
                    submitting: false,
                    errors: {
                        approver: data.errors.approver_id,
                        description: data.errors.description,
                    },
                }));
            } else {
                // Generic server error
                setForm(f => ({
                    ...f,
                    submitting: false,
                    errors: {
                        description: data.error || 'An error occurred. Please try again.',
                    },
                }));
            }
        } catch (err) {
            // Network error - check mount state first
            if (!isMountedRef.current) {
                return;
            }
            setForm(f => ({
                ...f,
                submitting: false,
                errors: {
                    description: 'Network error. Please try again.',
                },
            }));
        }
    }, [validate, onClose, channelId, teamId, form.approverId, form.description]);

    /**
     * Task 3: Handle approver selection change
     */
    const handleApproverChange = useCallback((userId: string) => {
        setForm(f => ({
            ...f,
            approverId: userId,
            errors: {...f.errors, approver: undefined},
        }));
    }, []);

    /**
     * Task 3: Handle description change
     */
    const handleDescriptionChange = useCallback((text: string) => {
        setForm(f => ({
            ...f,
            description: text,
            errors: {...f.errors, description: undefined},
        }));
    }, []);

    /**
     * Task 6: Handle cancel button click
     */
    const handleCancel = useCallback(() => {
        onClose();
    }, [onClose]);

    /**
     * Get button style based on variant and state
     */
    const getButtonStyle = (variant: 'primary' | 'secondary', disabled: boolean): React.CSSProperties => {
        return {
            ...buttonStyles.base,
            ...(variant === 'primary' ? buttonStyles.primary : buttonStyles.secondary),
            ...(disabled ? buttonStyles.disabled : {}),
        };
    };

    return (
        <Modal
            visible={visible}
            onClose={onClose}
            title="Request Approval"
        >
            <div style={styles.form} data-testid="approval-request-form">
                {/* Approver selector */}
                <UserSelector
                    value={form.approverId}
                    onChange={handleApproverChange}
                    error={form.errors.approver}
                    label="Select Approver"
                    placeholder="Search for a user..."
                    excludeUserIds={[currentUserId]}
                    disabled={form.submitting}
                />

                {/* Description text area */}
                <TextArea
                    value={form.description}
                    onChange={handleDescriptionChange}
                    error={form.errors.description}
                    label="What needs approval?"
                    placeholder="Describe what needs approval..."
                    maxLength={1000}
                    disabled={form.submitting}
                    rows={4}
                />

                {/* Button row */}
                <div style={styles.buttonRow}>
                    <button
                        type="button"
                        onClick={handleCancel}
                        disabled={form.submitting}
                        style={getButtonStyle('secondary', form.submitting)}
                        data-testid="cancel-button"
                    >
                        Cancel
                    </button>
                    <button
                        type="button"
                        onClick={handleSubmit}
                        disabled={form.submitting}
                        style={getButtonStyle('primary', form.submitting)}
                        data-testid="submit-button"
                    >
                        {form.submitting ? 'Submitting...' : 'Submit'}
                    </button>
                </div>
            </div>
        </Modal>
    );
};

// Memoize for performance
const ApprovalRequestModal = React.memo(ApprovalRequestModalInner);

export default ApprovalRequestModal;
