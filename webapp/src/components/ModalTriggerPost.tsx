import React, {useEffect, useRef, useContext} from 'react';
import ModalContext, {dispatchModalOpen} from '../context/ModalContext';

/**
 * Post interface for ModalTriggerPost
 * Matches Mattermost post structure
 */
interface Post {
    id: string;
    props?: {
        modal_type?: string;
        [key: string]: any;
    } | null;
}

/**
 * ModalTriggerPost Props
 */
interface ModalTriggerPostProps {
    post: Post;
}

/**
 * ModalTriggerPost Component
 *
 * Story 11.1 - AC1: Custom Post Type for Modal Trigger
 *
 * An invisible component that triggers a modal to open when mounted.
 * This is registered as a custom post type (`custom_approval_modal`) and
 * used to trigger React modals from server-side ephemeral posts.
 *
 * Expected post.props format:
 * {
 *   modal_type: string;        // Type of modal to open (e.g., 'approval_request')
 *   channel_id?: string;       // Channel context
 *   team_id?: string;          // Team context
 *   trigger_user?: string;     // User who triggered the modal
 *   [key: string]: any;        // Additional props passed to modal
 * }
 *
 * Flow:
 * 1. Server creates ephemeral post with type 'custom_approval_modal'
 * 2. Mattermost renders this component for the post
 * 3. On mount, this component calls openModal() with the modal_type
 * 4. The ModalProvider renders the appropriate modal
 *
 * Note on ephemeral post cleanup (Issue 3):
 * Ephemeral posts are automatically managed by Mattermost and are only visible
 * to the target user. They naturally expire and don't persist. Since this
 * component renders as invisible (display: none), the post doesn't affect UI.
 * Explicit cleanup via API would add complexity without meaningful benefit.
 *
 * @example
 * // Server-side (Go):
 * post := &model.Post{
 *     Type: "custom_approval_modal",
 *     Props: map[string]interface{}{
 *         "modal_type":   "approval_request",
 *         "channel_id":   channelId,
 *         "team_id":      teamId,
 *     },
 * }
 * p.API.SendEphemeralPost(userId, post)
 */
const ModalTriggerPost: React.FC<ModalTriggerPostProps> = ({post}) => {
    // Issue 2 fix: Safely access context with fallback to global events
    const context = useContext(ModalContext);
    const hasTriggered = useRef(false);

    useEffect(() => {
        // Only trigger once to prevent flicker on re-renders
        if (hasTriggered.current) {
            return;
        }

        // Safely extract props
        const props = post?.props;
        if (!props) {
            return;
        }

        // Extract modal_type
        const modalType = props.modal_type;
        if (!modalType || typeof modalType !== 'string') {
            return;
        }

        // Extract all other props (excluding modal_type) to pass to the modal
        const {modal_type: _type, ...modalProps} = props;

        // Mark as triggered before opening to prevent double-trigger
        hasTriggered.current = true;

        // Issue 2 fix: Try context first, fall back to global events
        if (context) {
            context.openModal(modalType, modalProps);
        } else {
            // Fallback: dispatch global event for components outside React tree
            console.debug('ModalTriggerPost: Using global event fallback (context not available)');
            dispatchModalOpen(modalType, modalProps);
        }
    }, [post, context]);

    // Render as hidden element - no visible content
    // Using display: none to ensure it doesn't affect layout
    return (
        <div
            data-testid="modal-trigger-post"
            style={{display: 'none'}}
        />
    );
};

export default ModalTriggerPost;
