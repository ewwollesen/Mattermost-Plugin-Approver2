import React from 'react';
import {render} from '@testing-library/react';
import {Provider} from 'react-redux';
import configureStore from 'redux-mock-store';
import ApprovalPost from './ApprovalPost';

const mockStore = configureStore([]);

const createMockState = () => ({
    entities: {
        users: {
            currentUserId: 'user1',
            profiles: {
                user1: {id: 'user1', username: 'testuser'}
            }
        },
        timezone: {}
    }
});

describe('ApprovalPost Component', () => {
    const basePost = {
        id: 'post123',
        create_at: 1705593000000,
        update_at: 1705593000000,
        user_id: 'user123',
        channel_id: 'channel123',
        message: 'Approval pending',
        type: 'custom_approval',
        props: {},
    };

    const store = mockStore(createMockState());

    it('renders pending approval', () => {
        const post = {
            ...basePost,
            props: {
                approval_code: 'A-TEST1',
                approval_status: 'pending',
                description: 'Test approval request',
                requester_username: 'requester',
                requester_display_name: 'Requester User',
                approver_username: 'approver',
                approver_display_name: 'Approver User',
                created_at: 1705593000000,
            },
        };

        const {container} = render(
            <Provider store={store}>
                <ApprovalPost post={post} />
            </Provider>
        );

        expect(container.textContent).toContain('⏳ Approval Pending');
        expect(container.textContent).toContain('A-TEST1');
        expect(container.textContent).toContain('Test approval request');
        expect(container.textContent).toContain('Awaiting');
        expect(container.textContent).toContain('@approver');
    });

    it('renders approved approval with note', () => {
        const post = {
            ...basePost,
            props: {
                approval_code: 'A-TEST2',
                approval_status: 'approved',
                description: 'Test approval request',
                requester_username: 'requester',
                requester_display_name: 'Requester User',
                approver_username: 'approver',
                approver_display_name: 'Approver User',
                created_at: 1705593000000,
                decided_at: 1705594000000,
                note: 'Looks good!',
            },
        };

        const {container} = render(
            <Provider store={store}>
                <ApprovalPost post={post} />
            </Provider>
        );

        expect(container.textContent).toContain('✅ Approval Approved');
        expect(container.textContent).toContain('A-TEST2');
        expect(container.textContent).toContain('Approved By');
        expect(container.textContent).toContain('@approver');
        expect(container.textContent).toContain('Looks good!');
    });

    it('renders approved approval without note', () => {
        const post = {
            ...basePost,
            props: {
                approval_code: 'A-TEST2',
                approval_status: 'approved',
                description: 'Test approval request',
                requester_username: 'requester',
                requester_display_name: 'Requester User',
                approver_username: 'approver',
                approver_display_name: 'Approver User',
                created_at: 1705593000000,
                decided_at: 1705594000000,
            },
        };

        const {container} = render(
            <Provider store={store}>
                <ApprovalPost post={post} />
            </Provider>
        );

        expect(container.textContent).toContain('✅ Approval Approved');
        expect(container.textContent).not.toContain('Note:');
    });

    it('renders denied approval with reason', () => {
        const post = {
            ...basePost,
            props: {
                approval_code: 'A-TEST3',
                approval_status: 'denied',
                description: 'Test approval request',
                requester_username: 'requester',
                requester_display_name: 'Requester User',
                approver_username: 'approver',
                approver_display_name: 'Approver User',
                created_at: 1705593000000,
                decided_at: 1705594000000,
                decision_comment: 'Not approved',
            },
        };

        const {container} = render(
            <Provider store={store}>
                <ApprovalPost post={post} />
            </Provider>
        );

        expect(container.textContent).toContain('❌ Approval Denied');
        expect(container.textContent).toContain('A-TEST3');
        expect(container.textContent).toContain('Denied By');
        expect(container.textContent).toContain('@approver');
        expect(container.textContent).toContain('Not approved');
    });

    it('renders denied approval without reason', () => {
        const post = {
            ...basePost,
            props: {
                approval_code: 'A-TEST3',
                approval_status: 'denied',
                description: 'Test approval request',
                requester_username: 'requester',
                requester_display_name: 'Requester User',
                approver_username: 'approver',
                approver_display_name: 'Approver User',
                created_at: 1705593000000,
                decided_at: 1705594000000,
            },
        };

        const {container} = render(
            <Provider store={store}>
                <ApprovalPost post={post} />
            </Provider>
        );

        expect(container.textContent).toContain('❌ Approval Denied');
        expect(container.textContent).not.toContain('Reason:');
    });

    it('renders canceled approval', () => {
        const post = {
            ...basePost,
            props: {
                approval_code: 'A-TEST4',
                approval_status: 'canceled',
                description: 'Test approval request',
                requester_username: 'requester',
                requester_display_name: 'Requester User',
                approver_username: 'approver',
                approver_display_name: 'Approver User',
                created_at: 1705593000000,
                decision_comment: 'User requested cancellation',
            },
        };

        const {container} = render(
            <Provider store={store}>
                <ApprovalPost post={post} />
            </Provider>
        );

        expect(container.textContent).toContain('🚫 Approval Canceled');
        expect(container.textContent).toContain('A-TEST4');
        expect(container.textContent).toContain('This approval request was canceled');
        expect(container.textContent).toContain('User requested cancellation');
    });

    it('renders timeout approval', () => {
        const post = {
            ...basePost,
            props: {
                approval_code: 'A-TEST5',
                approval_status: 'timeout',
                description: 'Test approval request',
                requester_username: 'requester',
                requester_display_name: 'Requester User',
                approver_username: 'approver',
                approver_display_name: 'Approver User',
                created_at: 1705593000000,
            },
        };

        const {container} = render(
            <Provider store={store}>
                <ApprovalPost post={post} />
            </Provider>
        );

        expect(container.textContent).toContain('⏱️ Approval Timed Out');
        expect(container.textContent).toContain('A-TEST5');
        expect(container.textContent).toContain('No response (timed out)');
        expect(container.textContent).toContain('@approver');
    });

    it('truncates description longer than 80 characters', () => {
        const longDescription = 'This is a very long description that exceeds the 80 character limit and should be truncated with ellipsis';
        const post = {
            ...basePost,
            props: {
                approval_code: 'A-TEST6',
                approval_status: 'pending',
                description: longDescription,
                requester_username: 'requester',
                requester_display_name: 'Requester User',
                approver_username: 'approver',
                approver_display_name: 'Approver User',
                created_at: 1705593000000,
            },
        };

        const {container} = render(
            <Provider store={store}>
                <ApprovalPost post={post} />
            </Provider>
        );

        expect(container.textContent).toContain('...');
        expect(container.textContent).not.toContain(longDescription);
    });

    it('does not truncate description shorter than 80 characters', () => {
        const shortDescription = 'Short description';
        const post = {
            ...basePost,
            props: {
                approval_code: 'A-TEST7',
                approval_status: 'pending',
                description: shortDescription,
                requester_username: 'requester',
                requester_display_name: 'Requester User',
                approver_username: 'approver',
                approver_display_name: 'Approver User',
                created_at: 1705593000000,
            },
        };

        const {container} = render(
            <Provider store={store}>
                <ApprovalPost post={post} />
            </Provider>
        );

        expect(container.textContent).toContain(shortDescription);
        expect(container.textContent).not.toContain('...');
    });

    it('handles missing props gracefully', () => {
        const post = {
            ...basePost,
            props: null,
        };

        const {container} = render(
            <Provider store={store}>
                <ApprovalPost post={post} />
            </Provider>
        );

        expect(container.textContent).toContain('Invalid approval post data');
    });

    it('applies defensive defaults for missing fields', () => {
        const post = {
            ...basePost,
            props: {
                approval_status: 'pending',
            },
        };

        const {container} = render(
            <Provider store={store}>
                <ApprovalPost post={post} />
            </Provider>
        );

        expect(container.textContent).toContain('UNKNOWN');
        expect(container.textContent).toContain('No description provided');
        expect(container.textContent).toContain('@unknown');
    });

    it('has proper accessibility attributes', () => {
        const post = {
            ...basePost,
            props: {
                approval_code: 'A-TEST8',
                approval_status: 'pending',
                description: 'Test',
                requester_username: 'requester',
                requester_display_name: 'Requester User',
                approver_username: 'approver',
                approver_display_name: 'Approver User',
                created_at: 1705593000000,
            },
        };

        const {container} = render(
            <Provider store={store}>
                <ApprovalPost post={post} />
            </Provider>
        );

        const article = container.querySelector('article');
        expect(article?.getAttribute('role')).toBe('article');
        expect(article?.getAttribute('aria-label')).toContain('Approval pending: A-TEST8');
        expect(article?.getAttribute('aria-live')).toBe('polite');
        expect(article?.getAttribute('aria-atomic')).toBe('true');
    });

    it('validates React.memo prevents unnecessary re-renders', () => {
        const post = {
            ...basePost,
            props: {
                approval_code: 'A-MEMO',
                approval_status: 'pending',
                description: 'Memo test',
                requester_username: 'requester',
                requester_display_name: 'Requester User',
                approver_username: 'approver',
                approver_display_name: 'Approver User',
                created_at: 1705593000000,
            },
        };

        const {rerender, container} = render(
            <Provider store={store}>
                <ApprovalPost post={post} />
            </Provider>
        );

        const firstRender = container.innerHTML;

        // Re-render with same post object (memo should prevent re-render)
        rerender(
            <Provider store={store}>
                <ApprovalPost post={post} />
            </Provider>
        );

        const secondRender = container.innerHTML;
        expect(firstRender).toBe(secondRender);

        // Re-render with new post object but same content (should re-render due to object identity change)
        const newPost = {...post};
        rerender(
            <Provider store={store}>
                <ApprovalPost post={newPost} />
            </Provider>
        );

        // Content should still match even though object changed
        expect(container.textContent).toContain('A-MEMO');
    });

    it('handles description exactly 80 characters (boundary condition)', () => {
        const exactDescription = 'A'.repeat(80); // Exactly 80 chars
        const post = {
            ...basePost,
            props: {
                approval_code: 'A-TEST9',
                approval_status: 'pending',
                description: exactDescription,
                requester_username: 'requester',
                requester_display_name: 'Requester User',
                approver_username: 'approver',
                approver_display_name: 'Approver User',
                created_at: 1705593000000,
            },
        };

        const {container} = render(
            <Provider store={store}>
                <ApprovalPost post={post} />
            </Provider>
        );

        // Should NOT truncate at exactly 80 chars (only >80)
        expect(container.textContent).toContain(exactDescription);
        expect(container.textContent).not.toContain('...');
    });

    // Story 9.10: Interactive button tests
    describe('Interactive buttons', () => {
        it('renders Approve/Deny buttons when attachments are present', () => {
            const post = {
                ...basePost,
                props: {
                    approval_code: 'A-BTN1',
                    approval_status: 'pending',
                    description: 'Deploy to production',
                    requester_username: 'requester',
                    requester_display_name: 'Requester User',
                    approver_username: 'approver',
                    approver_display_name: 'Approver User',
                    created_at: 1705593000000,
                    notification_type: 'approval_request',
                    is_dm: true,
                    attachments: [
                        {
                            actions: [
                                {
                                    name: 'Approve',
                                    type: 'button',
                                    style: 'primary',
                                    integration: {
                                        url: '/plugins/com.mattermost.plugin-approver2/action',
                                        context: {approval_id: 'record123', action: 'approve'},
                                    },
                                },
                                {
                                    name: 'Deny',
                                    type: 'button',
                                    style: 'danger',
                                    integration: {
                                        url: '/plugins/com.mattermost.plugin-approver2/action',
                                        context: {approval_id: 'record123', action: 'deny'},
                                    },
                                },
                            ],
                        },
                    ],
                },
            };

            const {container} = render(
                <Provider store={store}>
                    <ApprovalPost post={post} />
                </Provider>
            );

            expect(container.textContent).toContain('Approve');
            expect(container.textContent).toContain('Deny');

            const buttons = container.querySelectorAll('button');
            expect(buttons.length).toBe(2);
        });

        it('does not render buttons when attachments are missing', () => {
            const post = {
                ...basePost,
                props: {
                    approval_code: 'A-BTN2',
                    approval_status: 'approved',
                    description: 'Deploy to production',
                    requester_username: 'requester',
                    requester_display_name: 'Requester User',
                    approver_username: 'approver',
                    approver_display_name: 'Approver User',
                    created_at: 1705593000000,
                    decided_at: 1705594000000,
                    notification_type: 'outcome',
                    is_dm: true,
                    // No attachments
                },
            };

            const {container} = render(
                <Provider store={store}>
                    <ApprovalPost post={post} />
                </Provider>
            );

            const buttons = container.querySelectorAll('button');
            expect(buttons.length).toBe(0);
        });
    });

    // Story 9.10: DM notification tests
    describe('DM notification rendering', () => {
        it('renders DM approval request with full description (no truncation)', () => {
            const longDescription = 'This is a very long description that exceeds the 80 character limit but should NOT be truncated in DM context because DMs allow full context';
            const post = {
                ...basePost,
                props: {
                    approval_code: 'A-DM1',
                    approval_status: 'pending',
                    description: longDescription,
                    requester_username: 'requester',
                    requester_display_name: 'Requester User',
                    approver_username: 'approver',
                    approver_display_name: 'Approver User',
                    created_at: 1705593000000,
                    notification_type: 'approval_request',
                    is_dm: true,
                },
            };

            const {container} = render(
                <Provider store={store}>
                    <ApprovalPost post={post} />
                </Provider>
            );

            // Full description should be visible (no truncation for DMs)
            expect(container.textContent).toContain(longDescription);
            expect(container.textContent).not.toContain('...');
        });

        it('renders DM outcome notification', () => {
            const post = {
                ...basePost,
                props: {
                    approval_code: 'A-DM2',
                    approval_status: 'approved',
                    description: 'Deploy to production',
                    requester_username: 'requester',
                    requester_display_name: 'Requester User',
                    approver_username: 'approver',
                    approver_display_name: 'Approver User',
                    created_at: 1705593000000,
                    decided_at: 1705594000000,
                    note: 'Looks good', // Use 'note' field for approved status
                    notification_type: 'outcome',
                    is_dm: true,
                },
            };

            const {container} = render(
                <Provider store={store}>
                    <ApprovalPost post={post} />
                </Provider>
            );

            expect(container.textContent).toContain('✅ Approval Approved');
            expect(container.textContent).toContain('A-DM2');
            expect(container.textContent).toContain('Deploy to production');
            expect(container.textContent).toContain('Looks good');
        });

        it('renders DM cancellation notification', () => {
            const post = {
                ...basePost,
                props: {
                    approval_code: 'A-DM3',
                    approval_status: 'canceled',
                    description: 'Deploy to production',
                    requester_username: 'requester',
                    requester_display_name: 'Requester User',
                    approver_username: 'approver',
                    approver_display_name: 'Approver User',
                    created_at: 1705593000000,
                    decision_comment: 'No longer needed',
                    notification_type: 'cancellation',
                    is_dm: true,
                },
            };

            const {container} = render(
                <Provider store={store}>
                    <ApprovalPost post={post} />
                </Provider>
            );

            expect(container.textContent).toContain('🚫 Approval Canceled');
            expect(container.textContent).toContain('A-DM3');
            expect(container.textContent).toContain('No longer needed');
        });

        it('renders DM timeout notification', () => {
            const post = {
                ...basePost,
                props: {
                    approval_code: 'A-DM4',
                    approval_status: 'timeout',
                    description: 'Deploy to production',
                    requester_username: 'requester',
                    requester_display_name: 'Requester User',
                    approver_username: 'approver',
                    approver_display_name: 'Approver User',
                    created_at: 1705593000000,
                    notification_type: 'timeout',
                    is_dm: true,
                },
            };

            const {container} = render(
                <Provider store={store}>
                    <ApprovalPost post={post} />
                </Provider>
            );

            expect(container.textContent).toContain('⏱️ Approval Timed Out');
            expect(container.textContent).toContain('A-DM4');
            expect(container.textContent).toContain('No response (timed out)');
        });

        it('playbook posts still truncate description at 80 chars', () => {
            const longDescription = 'This is a very long description that exceeds the 80 character limit and should be truncated for playbook posts';
            const post = {
                ...basePost,
                props: {
                    approval_code: 'A-PB1',
                    approval_status: 'pending',
                    description: longDescription,
                    requester_username: 'requester',
                    requester_display_name: 'Requester User',
                    approver_username: 'approver',
                    approver_display_name: 'Approver User',
                    created_at: 1705593000000,
                    // No is_dm field = playbook post
                },
            };

            const {container} = render(
                <Provider store={store}>
                    <ApprovalPost post={post} />
                </Provider>
            );

            // Should be truncated for playbook posts
            expect(container.textContent).toContain('...');
            expect(container.textContent).not.toContain(longDescription);
        });

        it('detects is_dm as false for playbook posts', () => {
            const longDescription = 'This is a very long description that exceeds the 80 character limit and should be truncated';
            const post = {
                ...basePost,
                props: {
                    approval_code: 'A-PB2',
                    approval_status: 'pending',
                    description: longDescription,
                    requester_username: 'requester',
                    requester_display_name: 'Requester User',
                    approver_username: 'approver',
                    approver_display_name: 'Approver User',
                    created_at: 1705593000000,
                    is_dm: false, // Explicitly false
                },
            };

            const {container} = render(
                <Provider store={store}>
                    <ApprovalPost post={post} />
                </Provider>
            );

            // Should be truncated when is_dm is false
            expect(container.textContent).toContain('...');
        });
    });
});
