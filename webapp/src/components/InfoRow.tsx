import React, { ReactNode } from 'react';

export interface InfoRowProps {
    label: string;
    value: string | ReactNode;
    /** Emoji or icon character to display before label (e.g., "🔍", "📝") */
    icon?: string;
    /** Whether to show colon after label. Default: true */
    showColon?: boolean;
}

const InfoRow: React.FC<InfoRowProps> = React.memo(({
    label,
    value,
    icon,
    showColon = true
}) => {
    return (
        <div style={{
            marginBottom: '6px',
            fontSize: '14px',
            display: 'flex',
            alignItems: 'baseline',
            gap: '8px'
        }}>
            {icon && <span>{icon}</span>}
            <span style={{ fontWeight: 600 }}>
                {label}{showColon ? ':' : ''}
            </span>
            <span style={{
                color: 'var(--center-channel-color, #3d3c40)',
                flex: 1
            }}>
                {value}
            </span>
        </div>
    );
});

InfoRow.displayName = 'InfoRow';

export default InfoRow;
