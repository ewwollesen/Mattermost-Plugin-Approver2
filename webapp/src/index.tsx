// Mattermost Plugin Webapp Entry Point
// Plugin: Approval Workflow (Approver2)
// Version: 3.0.0

import ApprovalPost from './components/ApprovalPost';
import ApprovalDMPost from './components/ApprovalDMPost';

// Webpack injected version
declare const PLUGIN_VERSION: string;

// Make this file a module
export {};

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

            // Debug logging (verbose logs use console.debug)
            console.log(`Approver Plugin Webapp v${version} Initialized`);
            console.debug('Registered custom post type: custom_approval');
            console.debug('Registered custom post type: custom_approval_dm');
            console.debug('ApprovalPost component registered for playbook approval posts');
            console.debug('ApprovalDMPost component registered for DM approval notifications');
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
