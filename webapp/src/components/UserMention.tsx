import React from 'react';

export interface UserMentionProps {
    username: string;
    displayName?: string;
}

const UserMention: React.FC<UserMentionProps> = React.memo(({
    username,
    displayName
}) => {
    const title = displayName ? `${displayName} (@${username})` : username;

    return (
        <span
            title={title}
            style={{
                color: 'var(--link-color, #1c58d9)',
                // Note: cursor:pointer styling without onClick is intentional
                // Awaiting Mattermost mention system integration in future story
                // Styled to match Mattermost mention appearance for consistency
                cursor: 'pointer',
                fontWeight: 500
            }}
        >
            @{username}
        </span>
    );
});

UserMention.displayName = 'UserMention';

export default UserMention;
