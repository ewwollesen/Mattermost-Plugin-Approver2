// Mattermost Plugin Webapp Entry Point
// Plugin: Approval Workflow (Approver2)
// Version: 3.0.0

import React from 'react';
import ApprovalPost from './components/ApprovalPost';
import ApprovalDMPost from './components/ApprovalDMPost';
import ModalTriggerPost from './components/ModalTriggerPost';
import ApprovalRequestModal from './components/ApprovalRequestModal';
import {ModalProvider, useModal} from './context/ModalContext';

// Webpack injected version
declare const PLUGIN_VERSION: string;

// Make this file a module
export {};

/**
 * Modal types supported by the plugin
 * Story 11.3: Add 'approval_request' modal type
 */
const MODAL_TYPES = {
    APPROVAL_REQUEST: 'approval_request',
} as const;

/**
 * ModalRenderer Component
 * Story 11.3 - Task 8: Renders the appropriate modal based on context state
 *
 * This component listens to the ModalContext and renders the correct modal
 * component based on the modalType. It must be inside ModalProvider.
 */
const ModalRenderer: React.FC = () => {
    const {state, closeModal} = useModal();

    if (!state.isOpen || !state.modalType) {
        return null;
    }

    switch (state.modalType) {
        case MODAL_TYPES.APPROVAL_REQUEST:
            return (
                <ApprovalRequestModal
                    visible={true}
                    onClose={closeModal}
                    channelId={state.modalProps.channel_id || ''}
                    teamId={state.modalProps.team_id || ''}
                    currentUserId={state.modalProps.trigger_user || ''}
                />
            );
        default:
            console.warn(`Unknown modal type: ${state.modalType}`);
            return null;
    }
};

/**
 * RootComponent wrapping ModalProvider and ModalRenderer
 * This is registered as a root component to enable modal functionality
 */
const PluginRootComponent: React.FC = () => {
    return (
        <ModalProvider>
            <ModalRenderer />
        </ModalProvider>
    );
};

// Mattermost Plugin Registry Interface
interface PluginRegistry {
    registerPostTypeComponent(type: string, component: React.ComponentType<any>): void;
    registerRootComponent?(component: React.ComponentType): void;
    unregisterComponent?(componentId: string): void;
}

// Extend window interface for Mattermost plugin registration
declare global {
    interface Window {
        registerPlugin: (id: string, plugin: ApproverPlugin) => void;
    }
}

/**
 * Approver Plugin Webapp Entry Point
 * Registers custom post type for approval messages in Mattermost
 */
export class ApproverPlugin {
    /**
     * Called by Mattermost when plugin loads
     * @param registry - Mattermost plugin registry for component registration
     * @param store - Redux store (unused in this plugin)
     * @returns Cleanup function to unregister components on plugin unload
     */
    // eslint-disable-next-line @typescript-eslint/no-unused-vars, @typescript-eslint/no-explicit-any
    initialize(registry: PluginRegistry, store: any) {
        // Get plugin version with fallback
        const version = typeof PLUGIN_VERSION !== 'undefined' ? PLUGIN_VERSION : 'unknown';

        try {
            // Validate registry object
            if (!registry || typeof registry.registerPostTypeComponent !== 'function') {
                console.error('Approver Plugin: Invalid registry object - cannot register components');
                return;
            }

            // Register custom post type for approval messages (playbook channel posts)
            registry.registerPostTypeComponent('custom_approval', ApprovalPost);

            // Story 10.4: Register custom post type for DM notifications
            registry.registerPostTypeComponent('custom_approval_dm', ApprovalDMPost);

            // Story 11.1: Register custom post type for modal trigger
            // This invisible component opens React modals when server sends ephemeral post
            registry.registerPostTypeComponent('custom_approval_modal', ModalTriggerPost);

            // Story 11.1 & 11.3: Register root component for ModalProvider and ModalRenderer
            // PluginRootComponent includes ModalProvider + ModalRenderer for full modal support
            if (typeof registry.registerRootComponent === 'function') {
                registry.registerRootComponent(PluginRootComponent);
                console.debug('PluginRootComponent registered (ModalProvider + ModalRenderer)');
            } else {
                console.debug('registerRootComponent not available - modal trigger may not work');
            }

            // Debug logging (verbose logs use console.debug)
            console.log(`Approver Plugin Webapp v${version} Initialized`);
            console.debug('Registered custom post type: custom_approval');
            console.debug('Registered custom post type: custom_approval_dm');
            console.debug('Registered custom post type: custom_approval_modal');
            console.debug('ApprovalPost component registered for playbook approval posts');
            console.debug('ApprovalDMPost component registered for DM approval notifications');
            console.debug('ModalTriggerPost component registered for modal triggers');
        } catch (error) {
            console.error('Approver Plugin: Failed to register custom post type', error);
            return;
        }

        // Return cleanup function for proper plugin unloading
        return () => {
            console.debug('Approver Plugin: Cleanup completed');
        };
    }
}

// Register plugin with Mattermost
if (typeof window !== 'undefined' && window.registerPlugin) {
    window.registerPlugin('com.mattermost.plugin-approver2', new ApproverPlugin());
}
