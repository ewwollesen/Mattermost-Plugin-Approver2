import React from 'react';
import {render, screen, act} from '@testing-library/react';
import {ModalProvider, useModal} from '../context/ModalContext';
import ModalTriggerPost from './ModalTriggerPost';

// Helper component to check modal state
const ModalStateChecker: React.FC = () => {
    const {state} = useModal();
    return (
        <div>
            <span data-testid="modal-is-open">{state.isOpen.toString()}</span>
            <span data-testid="modal-type">{state.modalType || 'null'}</span>
            <span data-testid="modal-props">{JSON.stringify(state.modalProps)}</span>
        </div>
    );
};

describe('ModalTriggerPost Component (AC1)', () => {
    const basePost = {
        id: 'post123',
        create_at: 1705593000000,
        update_at: 1705593000000,
        user_id: 'user123',
        channel_id: 'channel123',
        message: '',
        type: 'custom_approval_modal',
        props: {
            modal_type: 'approval_request',
            channel_id: 'channel456',
            team_id: 'team789',
            trigger_user: 'user999',
        },
    };

    describe('Render behavior (AC1)', () => {
        it('renders as empty/invisible container', () => {
            const {container} = render(
                <ModalProvider>
                    <ModalTriggerPost post={basePost} />
                </ModalProvider>
            );

            // Should render an empty div with no visible content
            const triggerDiv = container.querySelector('[data-testid="modal-trigger-post"]');
            expect(triggerDiv).toBeInTheDocument();
            expect(triggerDiv?.textContent).toBe('');
        });

        it('has hidden styling to not affect layout', () => {
            const {container} = render(
                <ModalProvider>
                    <ModalTriggerPost post={basePost} />
                </ModalProvider>
            );

            const triggerDiv = container.querySelector('[data-testid="modal-trigger-post"]');
            expect(triggerDiv).toHaveStyle({display: 'none'});
        });
    });

    describe('Modal trigger behavior (AC1)', () => {
        it('opens modal on mount with props from post', () => {
            render(
                <ModalProvider>
                    <ModalStateChecker />
                    <ModalTriggerPost post={basePost} />
                </ModalProvider>
            );

            expect(screen.getByTestId('modal-is-open').textContent).toBe('true');
            expect(screen.getByTestId('modal-type').textContent).toBe('approval_request');

            const propsJson = screen.getByTestId('modal-props').textContent;
            const parsedProps = JSON.parse(propsJson || '{}');
            expect(parsedProps.channel_id).toBe('channel456');
            expect(parsedProps.team_id).toBe('team789');
            expect(parsedProps.trigger_user).toBe('user999');
        });

        it('extracts modal_type from post.props', () => {
            const postWithDifferentType = {
                ...basePost,
                props: {
                    modal_type: 'confirmation_dialog',
                    custom_data: {foo: 'bar'},
                },
            };

            render(
                <ModalProvider>
                    <ModalStateChecker />
                    <ModalTriggerPost post={postWithDifferentType} />
                </ModalProvider>
            );

            expect(screen.getByTestId('modal-type').textContent).toBe('confirmation_dialog');
        });

        it('only triggers once even if re-rendered', () => {
            const {rerender} = render(
                <ModalProvider>
                    <ModalStateChecker />
                    <ModalTriggerPost post={basePost} />
                </ModalProvider>
            );

            expect(screen.getByTestId('modal-is-open').textContent).toBe('true');

            // Re-render with same post
            rerender(
                <ModalProvider>
                    <ModalStateChecker />
                    <ModalTriggerPost post={basePost} />
                </ModalProvider>
            );

            // Should still be open (not closed and re-opened causing flicker)
            expect(screen.getByTestId('modal-is-open').textContent).toBe('true');
        });
    });

    describe('Edge cases', () => {
        it('handles missing modal_type gracefully', () => {
            const postNoType = {
                ...basePost,
                props: {
                    channel_id: 'channel456',
                    // No modal_type
                },
            };

            render(
                <ModalProvider>
                    <ModalStateChecker />
                    <ModalTriggerPost post={postNoType} />
                </ModalProvider>
            );

            // Should not open modal without type
            expect(screen.getByTestId('modal-is-open').textContent).toBe('false');
        });

        it('handles null props gracefully', () => {
            const postNullProps = {
                ...basePost,
                props: null,
            };

            render(
                <ModalProvider>
                    <ModalStateChecker />
                    <ModalTriggerPost post={postNullProps} />
                </ModalProvider>
            );

            // Should not crash, modal should not open
            expect(screen.getByTestId('modal-is-open').textContent).toBe('false');
        });

        it('handles empty props gracefully', () => {
            const postEmptyProps = {
                ...basePost,
                props: {},
            };

            render(
                <ModalProvider>
                    <ModalStateChecker />
                    <ModalTriggerPost post={postEmptyProps} />
                </ModalProvider>
            );

            // Should not crash, modal should not open
            expect(screen.getByTestId('modal-is-open').textContent).toBe('false');
        });
    });

    describe('Props passing', () => {
        it('passes all post.props except modal_type to modal', () => {
            const postWithManyProps = {
                ...basePost,
                props: {
                    modal_type: 'approval_request',
                    channel_id: 'ch123',
                    team_id: 'tm456',
                    trigger_user: 'usr789',
                    custom_field: 'custom_value',
                },
            };

            render(
                <ModalProvider>
                    <ModalStateChecker />
                    <ModalTriggerPost post={postWithManyProps} />
                </ModalProvider>
            );

            const propsJson = screen.getByTestId('modal-props').textContent;
            const parsedProps = JSON.parse(propsJson || '{}');

            // modal_type should NOT be in modalProps (it's used as the type)
            expect(parsedProps.modal_type).toBeUndefined();

            // All other props should be passed through
            expect(parsedProps.channel_id).toBe('ch123');
            expect(parsedProps.team_id).toBe('tm456');
            expect(parsedProps.trigger_user).toBe('usr789');
            expect(parsedProps.custom_field).toBe('custom_value');
        });
    });

    describe('Global event fallback (Issue 2 fix)', () => {
        it('uses global event dispatch when rendered outside ModalProvider', () => {
            // Spy on console.debug to verify fallback is used
            const consoleSpy = jest.spyOn(console, 'debug').mockImplementation();

            // Mock window event listener to capture dispatched event
            let capturedEvent: CustomEvent<{type: string; props: Record<string, any>}> | null = null;
            const eventListener = (e: Event) => {
                capturedEvent = e as CustomEvent<{type: string; props: Record<string, any>}>;
            };
            window.addEventListener('approver-modal-open', eventListener);

            // Render WITHOUT ModalProvider
            render(<ModalTriggerPost post={basePost} />);

            // Verify fallback was used
            expect(consoleSpy).toHaveBeenCalledWith(
                'ModalTriggerPost: Using global event fallback (context not available)'
            );

            // Verify event was dispatched
            expect(capturedEvent).not.toBeNull();
            expect(capturedEvent!.detail.type).toBe('approval_request');
            expect(capturedEvent!.detail.props.channel_id).toBe('channel456');

            // Cleanup
            window.removeEventListener('approver-modal-open', eventListener);
            consoleSpy.mockRestore();
        });

        it('prefers context over global events when ModalProvider is available', () => {
            const consoleSpy = jest.spyOn(console, 'debug').mockImplementation();

            render(
                <ModalProvider>
                    <ModalStateChecker />
                    <ModalTriggerPost post={basePost} />
                </ModalProvider>
            );

            // Should NOT log fallback message
            expect(consoleSpy).not.toHaveBeenCalledWith(
                'ModalTriggerPost: Using global event fallback (context not available)'
            );

            // Modal should still open via context
            expect(screen.getByTestId('modal-is-open').textContent).toBe('true');

            consoleSpy.mockRestore();
        });
    });
});
