import React, {useMemo, useCallback} from 'react';
import {useSelector} from 'react-redux';
import {getUserTimezone} from 'mattermost-redux/selectors/entities/timezone';
import {getCurrentUser} from 'mattermost-redux/selectors/entities/users';
import {GlobalState} from 'mattermost-redux/types/store';
import moment from 'moment-timezone';

interface TimestampProps {
    unixMillis: number;
    format?: string;  // Default: 'lll'
    relative?: boolean;
}

const Timestamp: React.FC<TimestampProps> = React.memo(({
    unixMillis,
    format = 'lll',
    relative = false
}) => {
    const currentUser = useSelector(getCurrentUser);
    const userTimezoneData = useSelector((state: GlobalState) =>
        currentUser ? getUserTimezone(state, currentUser.id) : null
    );

    // Extract timezone resolution to avoid duplication
    const resolveTimezone = useCallback(() => {
        let tz = moment.tz.guess();
        if (userTimezoneData) {
            if (userTimezoneData.useAutomaticTimezone && userTimezoneData.automaticTimezone) {
                tz = userTimezoneData.automaticTimezone;
            } else if (userTimezoneData.manualTimezone) {
                tz = userTimezoneData.manualTimezone;
            }
        }
        return tz;
    }, [userTimezoneData]);

    const formattedTime = useMemo(() => {
        // Handle edge cases
        if (isNaN(unixMillis) || unixMillis < 0) {
            return 'Invalid date';
        }

        if (!unixMillis || unixMillis === 0) {
            return 'Not yet decided';
        }

        const tz = resolveTimezone();
        const momentObj = moment.tz(unixMillis, tz);

        if (relative) {
            return momentObj.fromNow();
        }

        return momentObj.format(format);
    }, [unixMillis, resolveTimezone, format, relative]);

    const fullTimestamp = useMemo(() => {
        if (isNaN(unixMillis) || !unixMillis || unixMillis === 0) {
            return '';
        }

        const tz = resolveTimezone();
        return moment.tz(unixMillis, tz).format('LLLL z');
    }, [unixMillis, resolveTimezone]);

    return (
        <span title={fullTimestamp}>
            {formattedTime}
        </span>
    );
});

Timestamp.displayName = 'Timestamp';

export default Timestamp;
