/**
 * Tests for ApprovalDMPost Component
 * Story 10.4: Webapp Component for DM Notifications
 */

import React from 'react';
import {render, screen, fireEvent, waitFor} from '@testing-library/react';
import {Provider} from 'react-redux';
import configureStore from 'redux-mock-store';
import ApprovalDMPost from './ApprovalDMPost';

// Mock mattermost-redux doPostAction
jest.mock('mattermost-redux/actions/posts', () => ({
    doPostAction: jest.fn((postId: string, actionId: string) => ({
        type: 'DO_POST_ACTION',
        postId,
        actionId,
    })),
}));

import {doPostAction} from 'mattermost-redux/actions/posts';

const mockStore = configureStore([]);

// Mock state with users for Timestamp component
const mockStoreState = {
    entities: {
        users: {
            currentUserId: 'user123',
            profiles: {
                user123: {
                    id: 'user123',
                    username: 'testuser',
                    timezone: {
                        useAutomaticTimezone: true,
                        automaticTimezone: 'America/New_York',
                        manualTimezone: '',
                    },
                },
            },
        },
        general: {
            config: {},
        },
        preferences: {
            myPreferences: {},
        },
        timezone: {
            timezones: [],
        },
    },
};

describe('ApprovalDMPost', () => {
    let store: ReturnType<typeof mockStore>;

    beforeEach(() => {
        store = mockStore(mockStoreState);
        jest.clearAllMocks();
    });

    // Helper to render component with Redux provider
    const renderWithProvider = (post: any) => {
        return render(
            <Provider store={store}>
                <ApprovalDMPost post={post} />
            </Provider>
        );
    };

    describe('AC2: Extract data from post.props', () => {
        it('should render approval request notification correctly', () => {
            const post = {
                id: 'post123',
                props: {
                    approval_code: 'A-TEST01',
                    approval_status: 'pending',
                    description: 'Please approve this request',
                    requester_username: 'alice',
                    requester_display_name: 'Alice Smith',
                    approver_username: 'bob',
                    approver_display_name: 'Bob Jones',
                    created_at: 1705680000000,
                    notification_type: 'approval_request',
                    is_dm: true,
                },
            };

            renderWithProvider(post);

            expect(screen.getByText('A-TEST01')).toBeInTheDocument();
            expect(screen.getByText('Please approve this request')).toBeInTheDocument();
            expect(screen.getByText('@alice')).toBeInTheDocument();
        });

        it('should handle missing props gracefully', () => {
            const post = {
                id: 'post123',
                props: null,
            };

            renderWithProvider(post);

            expect(screen.getByText('Invalid approval DM data')).toBeInTheDocument();
        });

        it('should use default values for missing fields', () => {
            const post = {
                id: 'post123',
                props: {
                    notification_type: 'approval_request',
                },
            };

            renderWithProvider(post);

            expect(screen.getByText('UNKNOWN')).toBeInTheDocument();
            expect(screen.getByText('No description provided')).toBeInTheDocument();
        });
    });

    describe('AC3: Notification Type Rendering', () => {
        it('should render approval_request type with requester info', () => {
            const post = {
                id: 'post123',
                props: {
                    approval_code: 'A-TEST01',
                    approval_status: 'pending',
                    description: 'Test request',
                    requester_username: 'alice',
                    requester_display_name: 'Alice Smith',
                    approver_username: 'bob',
                    approver_display_name: 'Bob Jones',
                    created_at: 1705680000000,
                    notification_type: 'approval_request',
                    is_dm: true,
                },
            };

            renderWithProvider(post);

            // InfoRow renders label with colon, so use regex
            expect(screen.getByText(/Requested By/)).toBeInTheDocument();
            expect(screen.getByText('@alice')).toBeInTheDocument();
        });

        it('should render outcome type with decision details', () => {
            const post = {
                id: 'post123',
                props: {
                    approval_code: 'A-TEST01',
                    approval_status: 'approved',
                    description: 'Test request',
                    requester_username: 'alice',
                    requester_display_name: 'Alice Smith',
                    approver_username: 'bob',
                    approver_display_name: 'Bob Jones',
                    created_at: 1705680000000,
                    decided_at: 1705690000000,
                    decision_comment: 'Looks good!',
                    notification_type: 'outcome',
                    is_dm: true,
                },
            };

            renderWithProvider(post);

            expect(screen.getByText(/Approved By/)).toBeInTheDocument();
            expect(screen.getByText('@bob')).toBeInTheDocument();
            expect(screen.getByText(/Note/)).toBeInTheDocument();
            expect(screen.getByText('Looks good!')).toBeInTheDocument();
        });

        it('should render denied outcome with reason', () => {
            const post = {
                id: 'post123',
                props: {
                    approval_code: 'A-TEST01',
                    approval_status: 'denied',
                    description: 'Test request',
                    requester_username: 'alice',
                    requester_display_name: 'Alice Smith',
                    approver_username: 'bob',
                    approver_display_name: 'Bob Jones',
                    created_at: 1705680000000,
                    decided_at: 1705690000000,
                    decision_comment: 'Not approved',
                    notification_type: 'outcome',
                    is_dm: true,
                },
            };

            renderWithProvider(post);

            expect(screen.getByText(/Denied By/)).toBeInTheDocument();
            expect(screen.getByText(/Reason/)).toBeInTheDocument();
            expect(screen.getByText('Not approved')).toBeInTheDocument();
        });

        it('should render cancellation type', () => {
            const post = {
                id: 'post123',
                props: {
                    approval_code: 'A-TEST01',
                    approval_status: 'canceled',
                    description: 'Test request',
                    requester_username: 'alice',
                    requester_display_name: 'Alice Smith',
                    approver_username: 'bob',
                    approver_display_name: 'Bob Jones',
                    created_at: 1705680000000,
                    // Story 10.6: Use canceled_reason instead of decision_comment
                    canceled_reason: 'No longer needed',
                    canceled_at: 1705690000000,
                    notification_type: 'cancellation',
                    is_dm: true,
                },
            };

            renderWithProvider(post);

            expect(screen.getByText('This approval request was canceled')).toBeInTheDocument();
            expect(screen.getByText('No longer needed')).toBeInTheDocument();
        });

        // Story 10.6: New test for cancellation with timestamp
        it('should render cancellation type with timestamp', () => {
            const post = {
                id: 'post123',
                props: {
                    approval_code: 'A-TEST01',
                    approval_status: 'canceled',
                    description: 'Test request',
                    requester_username: 'alice',
                    requester_display_name: 'Alice Smith',
                    approver_username: 'bob',
                    approver_display_name: 'Bob Jones',
                    created_at: 1705680000000,
                    canceled_reason: 'Budget constraints',
                    canceled_at: 1705690000000, // Jan 19, 2024 afternoon UTC
                    notification_type: 'cancellation',
                    is_dm: true,
                },
            };

            renderWithProvider(post);

            // Verify cancellation reason displays
            expect(screen.getByText('Budget constraints')).toBeInTheDocument();
            // Verify "Canceled At" label appears (timestamp rendered by Timestamp component)
            expect(screen.getByText('Canceled At:')).toBeInTheDocument();
        });

        // Story 10.6: requester_cancellation - sent to requester when approver cancels
        it('should render requester_cancellation type with approver info', () => {
            const post = {
                id: 'post123',
                props: {
                    approval_code: 'A-TEST01',
                    approval_status: 'canceled',
                    description: 'Test request',
                    requester_username: 'alice',
                    requester_display_name: 'Alice Smith',
                    approver_username: 'bob',
                    approver_display_name: 'Bob Jones',
                    created_at: 1705680000000,
                    canceled_reason: 'Not needed anymore',
                    canceled_at: 1705690000000,
                    notification_type: 'requester_cancellation',
                    is_dm: true,
                },
            };

            renderWithProvider(post);

            // Verify approver info is shown (not requester)
            expect(screen.getByText(/Canceled By/)).toBeInTheDocument();
            expect(screen.getByText('@bob')).toBeInTheDocument();
            // Verify status message specific to requester view
            expect(screen.getByText('This approval request was canceled by the approver')).toBeInTheDocument();
            // Verify cancellation reason
            expect(screen.getByText('Not needed anymore')).toBeInTheDocument();
        });

        // Story 10.6 Code Review: Edge case - cancellation without reason
        it('should render cancellation type without reason when not provided', () => {
            const post = {
                id: 'post123',
                props: {
                    approval_code: 'A-TEST01',
                    approval_status: 'canceled',
                    description: 'Test request',
                    requester_username: 'alice',
                    requester_display_name: 'Alice Smith',
                    approver_username: 'bob',
                    approver_display_name: 'Bob Jones',
                    created_at: 1705680000000,
                    // canceled_reason intentionally omitted
                    canceled_at: 1705690000000,
                    notification_type: 'cancellation',
                    is_dm: true,
                },
            };

            renderWithProvider(post);

            // Verify cancellation renders without crashing
            expect(screen.getByText('This approval request was canceled')).toBeInTheDocument();
            // Verify "Reason" label is NOT present when no reason provided
            expect(screen.queryByText('Reason:')).not.toBeInTheDocument();
        });

        // Story 10.6 Code Review: Edge case - cancellation without timestamp
        it('should render cancellation type without timestamp when not provided', () => {
            const post = {
                id: 'post123',
                props: {
                    approval_code: 'A-TEST01',
                    approval_status: 'canceled',
                    description: 'Test request',
                    requester_username: 'alice',
                    requester_display_name: 'Alice Smith',
                    approver_username: 'bob',
                    approver_display_name: 'Bob Jones',
                    created_at: 1705680000000,
                    canceled_reason: 'No longer needed',
                    // canceled_at intentionally omitted
                    notification_type: 'cancellation',
                    is_dm: true,
                },
            };

            renderWithProvider(post);

            // Verify cancellation renders without crashing
            expect(screen.getByText('This approval request was canceled')).toBeInTheDocument();
            expect(screen.getByText('No longer needed')).toBeInTheDocument();
            // Verify "Canceled At" label is NOT present when no timestamp provided
            expect(screen.queryByText('Canceled At:')).not.toBeInTheDocument();
        });

        it('should render timeout type', () => {
            const post = {
                id: 'post123',
                props: {
                    approval_code: 'A-TEST01',
                    approval_status: 'timeout',
                    description: 'Test request',
                    requester_username: 'alice',
                    requester_display_name: 'Alice Smith',
                    approver_username: 'bob',
                    approver_display_name: 'Bob Jones',
                    created_at: 1705680000000,
                    notification_type: 'timeout',
                    is_dm: true,
                },
            };

            renderWithProvider(post);

            expect(screen.getByText(/Approver/)).toBeInTheDocument();
            expect(screen.getByText('No response received (timed out)')).toBeInTheDocument();
        });

        // Story 10.8: Updated to use verified_at and verification_comment props
        it('should render verification type with timestamp and comment', () => {
            const post = {
                id: 'post123',
                props: {
                    approval_code: 'A-TEST01',
                    approval_status: 'approved',
                    description: 'Test request',
                    requester_username: 'alice',
                    requester_display_name: 'Alice Smith',
                    approver_username: 'bob',
                    approver_display_name: 'Bob Jones',
                    created_at: 1705680000000,
                    verified_at: 1705690000000, // Story 10.8: Use verified_at
                    verification_comment: 'Task completed', // Story 10.8: Use verification_comment
                    notification_type: 'verification',
                    is_dm: true,
                },
            };

            renderWithProvider(post);

            expect(screen.getByText(/Verified By/)).toBeInTheDocument();
            // M1 Fix: Verify "Verified At" timestamp label renders
            expect(screen.getByText('Verified At:')).toBeInTheDocument();
            expect(screen.getByText('Task completed')).toBeInTheDocument();
        });

        // M2 Fix: Test verification WITHOUT comment
        it('should render verification type without comment when not provided', () => {
            const post = {
                id: 'post123',
                props: {
                    approval_code: 'A-TEST01',
                    approval_status: 'approved',
                    description: 'Test request',
                    requester_username: 'alice',
                    requester_display_name: 'Alice Smith',
                    approver_username: 'bob',
                    approver_display_name: 'Bob Jones',
                    created_at: 1705680000000,
                    verified_at: 1705690000000,
                    // verification_comment intentionally omitted
                    notification_type: 'verification',
                    is_dm: true,
                },
            };

            renderWithProvider(post);

            expect(screen.getByText(/Verified By/)).toBeInTheDocument();
            expect(screen.getByText('Verified At:')).toBeInTheDocument();
            // Note label should NOT appear when no comment provided
            expect(screen.queryByText('Note:')).not.toBeInTheDocument();
        });

        // M3 Fix: Test verification WITHOUT timestamp (verified_at = 0)
        it('should render verification type without timestamp when verified_at is zero', () => {
            const post = {
                id: 'post123',
                props: {
                    approval_code: 'A-TEST01',
                    approval_status: 'approved',
                    description: 'Test request',
                    requester_username: 'alice',
                    requester_display_name: 'Alice Smith',
                    approver_username: 'bob',
                    approver_display_name: 'Bob Jones',
                    created_at: 1705680000000,
                    verified_at: 0, // Zero timestamp - should not render
                    verification_comment: 'Task completed',
                    notification_type: 'verification',
                    is_dm: true,
                },
            };

            renderWithProvider(post);

            expect(screen.getByText(/Verified By/)).toBeInTheDocument();
            // Verified At should NOT appear when timestamp is 0
            expect(screen.queryByText('Verified At:')).not.toBeInTheDocument();
            // Comment should still render
            expect(screen.getByText('Task completed')).toBeInTheDocument();
        });
    });

    describe('AC4: Button Rendering', () => {
        it('should render buttons for approval_request with pending status', () => {
            const post = {
                id: 'post123',
                props: {
                    approval_code: 'A-TEST01',
                    approval_status: 'pending',
                    description: 'Test request',
                    requester_username: 'alice',
                    requester_display_name: 'Alice Smith',
                    approver_username: 'bob',
                    approver_display_name: 'Bob Jones',
                    created_at: 1705680000000,
                    notification_type: 'approval_request',
                    is_dm: true,
                    attachments: [
                        {
                            actions: [
                                {
                                    id: 'approve',
                                    name: 'Approve',
                                    style: 'success',
                                    integration: {
                                        url: '/plugins/com.mattermost.plugin-approver2/api/v1/approval/A-TEST01/approve',
                                    },
                                },
                                {
                                    id: 'deny',
                                    name: 'Deny',
                                    style: 'danger',
                                    integration: {
                                        url: '/plugins/com.mattermost.plugin-approver2/api/v1/approval/A-TEST01/deny',
                                    },
                                },
                            ],
                        },
                    ],
                },
            };

            renderWithProvider(post);

            expect(screen.getByRole('button', {name: 'Approve'})).toBeInTheDocument();
            expect(screen.getByRole('button', {name: 'Deny'})).toBeInTheDocument();
        });

        it('should NOT render buttons for non-pending status', () => {
            const post = {
                id: 'post123',
                props: {
                    approval_code: 'A-TEST01',
                    approval_status: 'approved',
                    description: 'Test request',
                    requester_username: 'alice',
                    requester_display_name: 'Alice Smith',
                    approver_username: 'bob',
                    approver_display_name: 'Bob Jones',
                    created_at: 1705680000000,
                    notification_type: 'approval_request',
                    is_dm: true,
                    attachments: [
                        {
                            actions: [
                                {id: 'approve', name: 'Approve', style: 'success'},
                                {id: 'deny', name: 'Deny', style: 'danger'},
                            ],
                        },
                    ],
                },
            };

            renderWithProvider(post);

            expect(screen.queryByRole('button', {name: 'Approve'})).not.toBeInTheDocument();
            expect(screen.queryByRole('button', {name: 'Deny'})).not.toBeInTheDocument();
        });

        it('should NOT render buttons for outcome notification type', () => {
            const post = {
                id: 'post123',
                props: {
                    approval_code: 'A-TEST01',
                    approval_status: 'pending',
                    description: 'Test request',
                    requester_username: 'alice',
                    requester_display_name: 'Alice Smith',
                    approver_username: 'bob',
                    approver_display_name: 'Bob Jones',
                    created_at: 1705680000000,
                    notification_type: 'outcome',
                    is_dm: true,
                    attachments: [
                        {
                            actions: [
                                {id: 'approve', name: 'Approve', style: 'success'},
                                {id: 'deny', name: 'Deny', style: 'danger'},
                            ],
                        },
                    ],
                },
            };

            renderWithProvider(post);

            expect(screen.queryByRole('button', {name: 'Approve'})).not.toBeInTheDocument();
            expect(screen.queryByRole('button', {name: 'Deny'})).not.toBeInTheDocument();
        });

        it('should call doPostAction when button is clicked', async () => {
            const post = {
                id: 'post123',
                props: {
                    approval_code: 'A-TEST01',
                    approval_status: 'pending',
                    description: 'Test request',
                    requester_username: 'alice',
                    requester_display_name: 'Alice Smith',
                    approver_username: 'bob',
                    approver_display_name: 'Bob Jones',
                    created_at: 1705680000000,
                    notification_type: 'approval_request',
                    is_dm: true,
                    attachments: [
                        {
                            actions: [
                                {
                                    id: 'approve',
                                    name: 'Approve',
                                    style: 'success',
                                    integration: {
                                        url: '/plugins/com.mattermost.plugin-approver2/api/v1/approval/A-TEST01/approve',
                                    },
                                },
                            ],
                        },
                    ],
                },
            };

            renderWithProvider(post);

            const approveButton = screen.getByRole('button', {name: 'Approve'});
            fireEvent.click(approveButton);

            await waitFor(() => {
                expect(doPostAction).toHaveBeenCalledWith('post123', 'approve');
            });
        });
    });

    describe('AC5: Timestamp Rendering', () => {
        it('should display timestamps using Timestamp component', () => {
            const post = {
                id: 'post123',
                props: {
                    approval_code: 'A-TEST01',
                    approval_status: 'pending',
                    description: 'Test request',
                    requester_username: 'alice',
                    requester_display_name: 'Alice Smith',
                    approver_username: 'bob',
                    approver_display_name: 'Bob Jones',
                    created_at: 1705680000000,
                    notification_type: 'approval_request',
                    is_dm: true,
                },
            };

            renderWithProvider(post);

            // Verify the timestamp row exists with label "Requested:"
            expect(screen.getByText('Requested:')).toBeInTheDocument();
            // Timestamp will render the date in local timezone - it's a span, not time element
            const dateText = screen.getByText(/Jan 19, 2024/);
            expect(dateText).toBeInTheDocument();
        });
    });

    describe('AC6: Redux Integration', () => {
        it('should dispatch doPostAction through Redux', async () => {
            const post = {
                id: 'post123',
                props: {
                    approval_code: 'A-TEST01',
                    approval_status: 'pending',
                    description: 'Test request',
                    requester_username: 'alice',
                    requester_display_name: 'Alice Smith',
                    approver_username: 'bob',
                    approver_display_name: 'Bob Jones',
                    created_at: 1705680000000,
                    notification_type: 'approval_request',
                    is_dm: true,
                    attachments: [
                        {
                            actions: [
                                {
                                    id: 'deny',
                                    name: 'Deny',
                                    style: 'danger',
                                    integration: {
                                        url: '/plugins/com.mattermost.plugin-approver2/api/v1/approval/A-TEST01/deny',
                                    },
                                },
                            ],
                        },
                    ],
                },
            };

            renderWithProvider(post);

            const denyButton = screen.getByRole('button', {name: 'Deny'});
            fireEvent.click(denyButton);

            await waitFor(() => {
                const actions = store.getActions();
                expect(actions).toContainEqual({
                    type: 'DO_POST_ACTION',
                    postId: 'post123',
                    actionId: 'deny',
                });
            });
        });
    });

    describe('Button Styling', () => {
        it('should apply success style to Approve button', () => {
            const post = {
                id: 'post123',
                props: {
                    approval_code: 'A-TEST01',
                    approval_status: 'pending',
                    description: 'Test',
                    requester_username: 'alice',
                    requester_display_name: 'Alice',
                    approver_username: 'bob',
                    approver_display_name: 'Bob',
                    created_at: 1705680000000,
                    notification_type: 'approval_request',
                    is_dm: true,
                    attachments: [
                        {
                            actions: [
                                {id: 'approve', name: 'Approve', style: 'success'},
                            ],
                        },
                    ],
                },
            };

            renderWithProvider(post);

            const button = screen.getByRole('button', {name: 'Approve'});
            expect(button).toHaveStyle({backgroundColor: 'var(--button-bg, #339970)'});
        });

        it('should apply danger style to Deny button', () => {
            const post = {
                id: 'post123',
                props: {
                    approval_code: 'A-TEST01',
                    approval_status: 'pending',
                    description: 'Test',
                    requester_username: 'alice',
                    requester_display_name: 'Alice',
                    approver_username: 'bob',
                    approver_display_name: 'Bob',
                    created_at: 1705680000000,
                    notification_type: 'approval_request',
                    is_dm: true,
                    attachments: [
                        {
                            actions: [
                                {id: 'deny', name: 'Deny', style: 'danger'},
                            ],
                        },
                    ],
                },
            };

            renderWithProvider(post);

            const button = screen.getByRole('button', {name: 'Deny'});
            expect(button).toHaveStyle({backgroundColor: 'var(--error-text, #d24b4e)'});
        });
    });

    describe('Accessibility', () => {
        it('should have proper ARIA attributes', () => {
            const post = {
                id: 'post123',
                props: {
                    approval_code: 'A-TEST01',
                    approval_status: 'pending',
                    description: 'Test',
                    requester_username: 'alice',
                    requester_display_name: 'Alice',
                    approver_username: 'bob',
                    approver_display_name: 'Bob',
                    created_at: 1705680000000,
                    notification_type: 'approval_request',
                    is_dm: true,
                },
            };

            renderWithProvider(post);

            const article = screen.getByRole('article');
            expect(article).toHaveAttribute('aria-label', 'Approval DM: A-TEST01');
            expect(article).toHaveAttribute('aria-live', 'polite');
        });
    });
});
