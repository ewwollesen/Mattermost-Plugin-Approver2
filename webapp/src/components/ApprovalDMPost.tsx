/**
 * ApprovalDMPost Component
 *
 * Story 10.4: Webapp Component for DM Notifications
 *
 * Renders DM approval notifications as custom components with:
 * - Interactive Approve/Deny buttons using doPostAction
 * - Timezone-aware timestamps
 * - Support for all notification types
 *
 * Uses Matterpoll pattern: doPostAction(postId, actionId) for button clicks
 */

import React, {useMemo, useState, useCallback} from 'react';
import {connect} from 'react-redux';
import {bindActionCreators, Dispatch, AnyAction} from 'redux';
import {doPostAction} from 'mattermost-redux/actions/posts';
import {StatusBadge, UserMention, InfoRow, Timestamp} from './index';

// Notification types from Story 10.3
// Story 10.6: Added 'requester_cancellation' for when approver cancels (sent to requester)
type NotificationType = 'approval_request' | 'outcome' | 'cancellation' | 'requester_cancellation' | 'timeout' | 'verification';

// Status types
type ApprovalStatus = 'pending' | 'approved' | 'denied' | 'canceled' | 'timeout';

// Approval data extracted from post.props
interface ApprovalDMData {
    code: string;
    description: string;
    status: ApprovalStatus;
    requesterUsername: string;
    requesterDisplayName: string;
    approverUsername: string;
    approverDisplayName: string;
    createdAt: number;
    decidedAt?: number;
    decisionComment?: string;
    canceledAt?: number; // Story 10.6: Cancellation timestamp
    canceledReason?: string; // Story 10.6: Cancellation reason (distinct from decisionComment)
    verifiedAt?: number; // Story 10.8: Verification timestamp
    verificationComment?: string; // Story 10.8: Verification comment
    notificationType: NotificationType;
}

// Button action from attachment
interface ButtonAction {
    id: string;
    name: string;
    style?: 'success' | 'danger' | 'primary' | 'default';
    integration?: {
        url: string;
        context?: Record<string, unknown>;
    };
}

// Component props
export interface ApprovalDMPostProps {
    post: {
        id: string;
        props?: Record<string, unknown>;
    };
}

// Props from Redux mapDispatchToProps
// L1 Fix: doPostAction returns a thunk that resolves to various types depending on success/failure.
// Using 'any' here avoids complex union types while maintaining runtime safety via try/catch.
// eslint-disable-next-line @typescript-eslint/no-explicit-any
type DoPostActionFn = (postId: string, actionId: string, selectedOption?: string) => any;

interface DispatchProps {
    actions: {
        doPostAction: DoPostActionFn;
    };
}

type Props = ApprovalDMPostProps & DispatchProps;

/**
 * ApprovalDMPost - Renders DM approval notifications
 * AC2: Extract data from post.props
 * AC3: Render based on notification_type
 * AC5: Use Timestamp component for timezone-aware display
 */
const ApprovalDMPost: React.FC<Props> = React.memo(({post, actions}) => {
    const [isProcessing, setIsProcessing] = useState(false);
    const [actionError, setActionError] = useState<string | null>(null);

    // AC2: Extract approval data from post.props
    const approvalData: ApprovalDMData | null = useMemo(() => {
        if (!post.props) {
            return null;
        }

        return {
            code: (post.props.approval_code as string) || 'UNKNOWN',
            description: (post.props.description as string) || 'No description provided',
            status: (post.props.approval_status as ApprovalStatus) || 'pending',
            requesterUsername: (post.props.requester_username as string) || 'unknown',
            requesterDisplayName: (post.props.requester_display_name as string) || 'Unknown User',
            approverUsername: (post.props.approver_username as string) || 'unknown',
            approverDisplayName: (post.props.approver_display_name as string) || 'Unknown User',
            createdAt: (post.props.created_at as number) || 0,
            decidedAt: post.props.decided_at as number | undefined,
            decisionComment: post.props.decision_comment as string | undefined,
            canceledAt: post.props.canceled_at as number | undefined, // Story 10.6
            canceledReason: post.props.canceled_reason as string | undefined, // Story 10.6
            verifiedAt: post.props.verified_at as number | undefined, // Story 10.8
            verificationComment: post.props.verification_comment as string | undefined, // Story 10.8
            notificationType: (post.props.notification_type as NotificationType) || 'approval_request',
        };
    }, [post.props]);

    // AC2: Extract buttons from attachment.actions
    const buttons: ButtonAction[] | null = useMemo(() => {
        if (!post.props?.attachments || !Array.isArray(post.props.attachments)) {
            return null;
        }

        const attachments = post.props.attachments as Array<{actions?: ButtonAction[]}>;
        const attachment = attachments[0];
        if (!attachment?.actions || !Array.isArray(attachment.actions)) {
            return null;
        }

        return attachment.actions;
    }, [post.props?.attachments]);

    // AC4: Handle button click using doPostAction
    const handleButtonClick = useCallback(async (actionId: string) => {
        setIsProcessing(true);
        setActionError(null);

        try {
            // AC4, AC6: Use doPostAction from mattermost-redux
            await actions.doPostAction(post.id, actionId);
            // Success - Mattermost will handle the response and update the post
        } catch (error) {
            setActionError(error instanceof Error ? error.message : 'Action failed');
        } finally {
            setIsProcessing(false);
        }
    }, [actions, post.id]);

    if (!approvalData) {
        return <div>Invalid approval DM data</div>;
    }

    // AC3: Render based on notification_type
    return (
        <article
            style={{
                padding: '12px',
                borderLeft: '4px solid var(--center-channel-color, #3d3c40)',
                backgroundColor: 'var(--center-channel-bg, #ffffff)',
                borderRadius: '4px',
                marginBottom: '8px',
                maxWidth: '100%',
                wordWrap: 'break-word',
            }}
            role="article"
            aria-label={`Approval DM: ${approvalData.code}`}
            aria-live="polite"
        >
            {/* Status Badge */}
            <StatusBadge status={approvalData.status} />

            {/* Request ID */}
            <InfoRow label="Request ID" value={approvalData.code} />

            {/* Description - Full text for DMs */}
            <InfoRow label="Description" value={approvalData.description} />

            {/* AC3: Notification Type Rendering */}
            {renderNotificationContent(approvalData)}

            {/* AC4: Button Rendering - Only for approval_request with pending status */}
            {approvalData.notificationType === 'approval_request' &&
             approvalData.status === 'pending' &&
             buttons &&
             buttons.length > 0 && (
                <div style={{marginTop: '12px', display: 'flex', gap: '8px'}}>
                    {buttons.map((button) => (
                        <button
                            key={button.id}
                            onClick={() => handleButtonClick(button.id)}
                            disabled={isProcessing}
                            type="button"
                            style={getButtonStyle(button.style, isProcessing)}
                        >
                            {isProcessing ? 'Processing...' : button.name}
                        </button>
                    ))}
                </div>
            )}

            {/* Error message */}
            {actionError && (
                <div
                    style={{
                        marginTop: '8px',
                        padding: '8px',
                        backgroundColor: 'var(--error-text, #d24b4e)',
                        color: '#ffffff',
                        borderRadius: '4px',
                        fontSize: '13px',
                    }}
                >
                    {actionError}
                </div>
            )}
        </article>
    );
});

/**
 * AC3: Render notification-specific content
 */
function renderNotificationContent(data: ApprovalDMData): React.ReactNode {
    switch (data.notificationType) {
        // AC3: approval_request - Show request details + awaiting approver
        case 'approval_request':
            return (
                <>
                    <InfoRow
                        label="Requested By"
                        value={<UserMention username={data.requesterUsername} displayName={data.requesterDisplayName} />}
                    />
                    {data.status === 'pending' && (
                        <InfoRow
                            label="Requested"
                            value={<Timestamp unixMillis={data.createdAt} />}
                        />
                    )}
                </>
            );

        // AC3: outcome - Show decision details (approved/denied)
        case 'outcome':
            return (
                <>
                    <InfoRow
                        label={data.status === 'approved' ? 'Approved By' : 'Denied By'}
                        value={<UserMention username={data.approverUsername} displayName={data.approverDisplayName} />}
                    />
                    <InfoRow
                        label={data.status === 'approved' ? 'Approved At' : 'Denied At'}
                        value={<Timestamp unixMillis={data.decidedAt || 0} />}
                    />
                    {data.decisionComment && (
                        <InfoRow label={data.status === 'approved' ? 'Note' : 'Reason'} value={data.decisionComment} />
                    )}
                </>
            );

        // AC3: cancellation - Show cancellation notice (sent to approver when requester cancels)
        // Story 10.6: Use canceledReason (from canceled_reason prop) instead of decisionComment
        case 'cancellation':
            return (
                <>
                    <InfoRow
                        label="Requested By"
                        value={<UserMention username={data.requesterUsername} displayName={data.requesterDisplayName} />}
                    />
                    <InfoRow label="Status" value="This approval request was canceled" />
                    {data.canceledReason && (
                        <InfoRow label="Reason" value={data.canceledReason} />
                    )}
                    {data.canceledAt && data.canceledAt > 0 && (
                        <InfoRow
                            label="Canceled At"
                            value={<Timestamp unixMillis={data.canceledAt} />}
                        />
                    )}
                </>
            );

        // Story 10.6: requester_cancellation - Sent to requester when approver cancels
        // Shows approver info instead of requester info
        case 'requester_cancellation':
            return (
                <>
                    <InfoRow
                        label="Canceled By"
                        value={<UserMention username={data.approverUsername} displayName={data.approverDisplayName} />}
                    />
                    <InfoRow label="Status" value="This approval request was canceled by the approver" />
                    {data.canceledReason && (
                        <InfoRow label="Reason" value={data.canceledReason} />
                    )}
                    {data.canceledAt && data.canceledAt > 0 && (
                        <InfoRow
                            label="Canceled At"
                            value={<Timestamp unixMillis={data.canceledAt} />}
                        />
                    )}
                </>
            );

        // AC3: timeout - Show timeout notice
        case 'timeout':
            return (
                <>
                    <InfoRow
                        label="Approver"
                        value={<UserMention username={data.approverUsername} displayName={data.approverDisplayName} />}
                    />
                    <InfoRow label="Status" value="No response received (timed out)" />
                    {data.createdAt > 0 && (
                        <InfoRow
                            label="Requested"
                            value={<Timestamp unixMillis={data.createdAt} />}
                        />
                    )}
                </>
            );

        // AC3: verification - Show verification confirmation
        // Story 10.8: Uses verifiedAt and verificationComment props from server
        case 'verification':
            return (
                <>
                    <InfoRow
                        label="Verified By"
                        value={<UserMention username={data.requesterUsername} displayName={data.requesterDisplayName} />}
                    />
                    {data.verifiedAt && data.verifiedAt > 0 && (
                        <InfoRow
                            label="Verified At"
                            value={<Timestamp unixMillis={data.verifiedAt} />}
                        />
                    )}
                    {data.verificationComment && (
                        <InfoRow label="Note" value={data.verificationComment} />
                    )}
                </>
            );

        default:
            return null;
    }
}

/**
 * Get button styles based on style type
 * AC4: success = green, danger = red
 */
function getButtonStyle(style: string | undefined, isProcessing: boolean): React.CSSProperties {
    const baseStyle: React.CSSProperties = {
        padding: '8px 16px',
        borderRadius: '4px',
        border: 'none',
        cursor: isProcessing ? 'not-allowed' : 'pointer',
        fontSize: '14px',
        fontWeight: 600,
        opacity: isProcessing ? 0.6 : 1,
    };

    switch (style) {
        case 'success':
            return {
                ...baseStyle,
                backgroundColor: 'var(--button-bg, #339970)',
                color: '#ffffff',
            };
        case 'danger':
            return {
                ...baseStyle,
                backgroundColor: 'var(--error-text, #d24b4e)',
                color: '#ffffff',
            };
        case 'primary':
            return {
                ...baseStyle,
                backgroundColor: 'var(--button-bg, #1c58d9)',
                color: '#ffffff',
            };
        default:
            return {
                ...baseStyle,
                backgroundColor: 'var(--center-channel-color-16, #e0e0e0)',
                color: 'var(--center-channel-color, #3d3c40)',
            };
    }
}

ApprovalDMPost.displayName = 'ApprovalDMPost';

// AC6: Connect to Redux with doPostAction
const mapDispatchToProps = (dispatch: Dispatch<AnyAction>): DispatchProps => ({
    actions: bindActionCreators({
        doPostAction,
    }, dispatch),
});

export default connect(null, mapDispatchToProps)(ApprovalDMPost);
