/**
 * User API Functions
 * Story 11.2: User Selector Component - Task 2
 *
 * Provides user search functionality using Mattermost's autocomplete API.
 */

/**
 * User option for dropdown display
 */
export interface UserOption {
    id: string;
    username: string;
    displayName: string;
    avatarUrl?: string;
}

/**
 * Mattermost user from API response
 */
interface MattermostUser {
    id: string;
    username: string;
    first_name: string;
    last_name: string;
    nickname: string;
}

/**
 * Mattermost autocomplete API response
 */
interface AutocompleteResponse {
    users: MattermostUser[];
    out_of_channel?: MattermostUser[];
}

const MATTERMOST_API_BASE = '/api/v4';

/**
 * Get display name from user object
 * Priority: nickname > first + last name > username
 */
export const getDisplayName = (user: MattermostUser): string => {
    if (user.nickname) {
        return user.nickname;
    }
    if (user.first_name || user.last_name) {
        return `${user.first_name} ${user.last_name}`.trim();
    }
    return user.username;
};

/**
 * Search for users using Mattermost autocomplete API
 *
 * @param term - Search term (minimum 2 characters)
 * @returns Array of user options for dropdown
 */
export const searchUsers = async (term: string): Promise<UserOption[]> => {
    // Require minimum 2 characters before searching
    if (term.length < 2) {
        return [];
    }

    try {
        const response = await fetch(
            `${MATTERMOST_API_BASE}/users/autocomplete?term=${encodeURIComponent(term)}`,
            {
                method: 'GET',
                headers: {
                    'Content-Type': 'application/json',
                },
            }
        );

        if (!response.ok) {
            console.error('User search failed:', response.status);
            return [];
        }

        const data: AutocompleteResponse = await response.json();

        return data.users.map((user) => ({
            id: user.id,
            username: user.username,
            displayName: getDisplayName(user),
            avatarUrl: `${MATTERMOST_API_BASE}/users/${user.id}/image?_=${Date.now()}`,
        }));
    } catch (error) {
        console.error('User search error:', error);
        return [];
    }
};
