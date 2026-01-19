import React, {useMemo, useState} from 'react';
import {StatusBadge, UserMention, InfoRow, Timestamp} from './index';

export interface ApprovalPostData {
    code: string;
    description: string;
    status: 'pending' | 'approved' | 'denied' | 'canceled' | 'timeout';
    requesterUsername: string;
    requesterDisplayName: string;
    approverUsername: string;
    approverDisplayName: string;
    createdAt: number;
    decidedAt?: number;
    decisionComment?: string;
    note?: string;
    // Story 9.10: DM-specific fields
    notificationType?: 'approval_request' | 'outcome' | 'cancellation' | 'timeout' | 'verification';
    isDM?: boolean;
}

export interface ApprovalPostProps {
    post: any; // Mattermost Post type
}

const ApprovalPost: React.FC<ApprovalPostProps> = React.memo(({post}) => {
    const [isProcessing, setIsProcessing] = useState(false);
    const [actionError, setActionError] = useState<string | null>(null);

    // Extract approval data from post.props
    const approvalData: ApprovalPostData | null = useMemo(() => {
        if (!post.props) {
            return null;
        }

        // Defensive extraction with defaults
        return {
            code: post.props.approval_code || 'UNKNOWN',
            description: post.props.description || 'No description provided',
            status: post.props.approval_status || 'pending',
            requesterUsername: post.props.requester_username || 'unknown',
            requesterDisplayName: post.props.requester_display_name || 'Unknown User',
            approverUsername: post.props.approver_username || 'unknown',
            approverDisplayName: post.props.approver_display_name || 'Unknown User',
            createdAt: post.props.created_at || 0,
            decidedAt: post.props.decided_at,
            decisionComment: post.props.decision_comment,
            note: post.props.note,
            // Story 9.10: DM-specific fields
            notificationType: post.props.notification_type,
            isDM: post.props.is_dm === true,
        };
    }, [post.props]);

    // Extract interactive buttons from attachments (Story 9.10: AC4)
    // Note: Buttons only work with standard posts, not custom post types
    // Mattermost strips the "integration" field from custom post type actions
    const buttons = useMemo(() => {
        if (!post.props?.attachments || !Array.isArray(post.props.attachments)) {
            return null;
        }

        const attachment = post.props.attachments[0];
        if (!attachment?.actions || !Array.isArray(attachment.actions)) {
            return null;
        }

        return attachment.actions;
    }, [post.props]);

    if (!approvalData) {
        return <div>Invalid approval post data</div>;
    }

    // Story 9.10: DM posts show full description, playbook posts truncate to 80 chars
    const desc = approvalData.description || '';
    const displayDescription = approvalData.isDM
        ? desc // Full description for DMs (AC3)
        : (desc.length > 80 ? desc.substring(0, 80) + '...' : desc); // Truncate for playbook posts

    // Handle button click (Story 9.10: AC4 - Interactive buttons in standard posts)
    // Note: This is only used for standard posts since custom post types strip integration data
    const handleButtonClick = async (event: React.MouseEvent<HTMLButtonElement>, action: any) => {
        // Prevent default to avoid any Mattermost interference
        event.preventDefault();
        event.stopPropagation();

        if (!action.integration) {
            setActionError('Button configuration error - no integration data');
            return;
        }

        setIsProcessing(true);
        setActionError(null);

        try {
            // Build the PostActionIntegrationRequest payload
            const payload = {
                user_id: (window as any).getCurrentUserId?.() || '',
                channel_id: post.channel_id,
                team_id: post.team_id || '',
                post_id: post.id,
                trigger_id: '',
                type: '',
                data_source: '',
                context: action.integration.context,
            };

            // POST to the action URL (plugin endpoint)
            const response = await fetch(action.integration.url, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'X-Requested-With': 'XMLHttpRequest',
                },
                body: JSON.stringify(payload),
                credentials: 'same-origin',
            });

            if (!response.ok) {
                let errorText = 'Action failed';
                try {
                    const errorJson = await response.json();
                    errorText = errorJson.error || errorJson.message || errorText;
                } catch (e) {
                    errorText = await response.text() || errorText;
                }
                throw new Error(errorText);
            }

            // Parse response
            const result = await response.json();

            // Check for ephemeral message (could be displayed in UI)
            if (result.ephemeral_text) {
                // Future: show ephemeral message to user
            }

            // Success - reload after a short delay to show updated state
            setTimeout(() => {
                window.location.reload();
            }, 500);

        } catch (error) {
            setActionError(error instanceof Error ? error.message : 'Action failed');
            setIsProcessing(false);
        }
    };

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
            aria-label={`Approval ${approvalData.status}: ${approvalData.code}`}
            aria-live="polite"
            aria-atomic="true"
        >
            {/* Status Badge */}
            <StatusBadge status={approvalData.status} />

            {/* Request ID */}
            <InfoRow label="Request ID" value={approvalData.code} />

            {/* Description */}
            <InfoRow label="Description" value={displayDescription} />

            {/* Status-specific rendering */}
            {approvalData.status === 'pending' && (
                <>
                    <InfoRow
                        label="Awaiting"
                        value={<UserMention username={approvalData.approverUsername} displayName={approvalData.approverDisplayName} />}
                    />
                    <InfoRow
                        label="Requested"
                        value={<Timestamp unixMillis={approvalData.createdAt} />}
                    />
                </>
            )}

            {approvalData.status === 'approved' && (
                <>
                    <InfoRow
                        label="Approved By"
                        value={<UserMention username={approvalData.approverUsername} displayName={approvalData.approverDisplayName} />}
                    />
                    <InfoRow
                        label="Approved At"
                        value={<Timestamp unixMillis={approvalData.decidedAt || 0} />}
                    />
                    {approvalData.note && (
                        <InfoRow label="Note" value={approvalData.note} />
                    )}
                </>
            )}

            {approvalData.status === 'denied' && (
                <>
                    <InfoRow
                        label="Denied By"
                        value={<UserMention username={approvalData.approverUsername} displayName={approvalData.approverDisplayName} />}
                    />
                    <InfoRow
                        label="Denied At"
                        value={<Timestamp unixMillis={approvalData.decidedAt || 0} />}
                    />
                    {approvalData.decisionComment && (
                        <InfoRow label="Reason" value={approvalData.decisionComment} />
                    )}
                </>
            )}

            {approvalData.status === 'canceled' && (
                <>
                    <InfoRow label="Canceled" value="This approval request was canceled" />
                    {approvalData.decisionComment && (
                        <InfoRow label="Reason" value={approvalData.decisionComment} />
                    )}
                </>
            )}

            {approvalData.status === 'timeout' && (
                <>
                    <InfoRow
                        label="Approver"
                        value={<UserMention username={approvalData.approverUsername} displayName={approvalData.approverDisplayName} />}
                    />
                    <InfoRow label="Status" value="No response (timed out)" />
                </>
            )}

            {/* Story 9.10: AC4 - Render interactive buttons for approval requests */}
            {buttons && buttons.length > 0 && (
                <div style={{marginTop: '12px', display: 'flex', gap: '8px'}}>
                    {buttons.map((button: any, index: number) => (
                        <button
                            key={index}
                            onClick={(e) => handleButtonClick(e, button)}
                            disabled={isProcessing}
                            type="button"
                            style={{
                                padding: '8px 16px',
                                borderRadius: '4px',
                                border: 'none',
                                cursor: isProcessing ? 'not-allowed' : 'pointer',
                                fontSize: '14px',
                                fontWeight: 600,
                                backgroundColor: button.style === 'primary'
                                    ? 'var(--button-bg, #1c58d9)'
                                    : button.style === 'danger'
                                        ? 'var(--error-text, #d24b4e)'
                                        : 'var(--center-channel-color-16, #e0e0e0)',
                                color: button.style === 'primary' || button.style === 'danger'
                                    ? '#ffffff'
                                    : 'var(--center-channel-color, #3d3c40)',
                                opacity: isProcessing ? 0.6 : 1,
                            }}
                        >
                            {isProcessing ? 'Processing...' : button.name}
                        </button>
                    ))}
                </div>
            )}

            {/* Show error message if action failed */}
            {actionError && (
                <div
                    style={{
                        marginTop: '8px',
                        padding: '8px',
                        backgroundColor: 'var(--error-text-color, #d24b4e)',
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

ApprovalPost.displayName = 'ApprovalPost';

export default ApprovalPost;
