import React from 'react';

export interface StatusBadgeProps {
    status: 'pending' | 'approved' | 'denied' | 'canceled' | 'timeout';
}

const STATUS_CONFIG: Record<StatusBadgeProps['status'], { emoji: string; text: string }> = {
    pending: { emoji: '⏳', text: 'Approval Pending' },
    approved: { emoji: '✅', text: 'Approval Approved' },
    denied: { emoji: '❌', text: 'Approval Denied' },
    canceled: { emoji: '🚫', text: 'Approval Canceled' },
    timeout: { emoji: '⏱️', text: 'Approval Timed Out' },
};

const StatusBadge: React.FC<StatusBadgeProps> = React.memo(({ status }) => {
    const config = STATUS_CONFIG[status];

    // Defensive check for invalid status values from API
    if (!config) {
        return (
            <div style={{
                fontSize: '18px',
                fontWeight: 600,
                marginBottom: '8px',
                color: 'var(--center-channel-color, #3d3c40)'
            }} role="status">
                ⚠️ Unknown Status
            </div>
        );
    }

    return (
        <div style={{
            fontSize: '18px',
            fontWeight: 600,
            marginBottom: '8px',
            color: 'var(--center-channel-color, #3d3c40)'
        }} role="status">
            {config.emoji} {config.text}
        </div>
    );
});

StatusBadge.displayName = 'StatusBadge';

export default StatusBadge;
