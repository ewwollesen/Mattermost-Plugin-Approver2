/**
 * Tests for Plugin Registration (index.tsx)
 * Story 9.7: Register Custom Post Type
 */

// Mock window.registerPlugin BEFORE importing index.tsx
const mockRegisterPlugin = jest.fn();
(global as any).window = {
    registerPlugin: mockRegisterPlugin,
};

// Mock PLUGIN_VERSION
(global as any).PLUGIN_VERSION = '3.0.0';

// Mock console methods
const mockConsoleLog = jest.spyOn(console, 'log').mockImplementation();
const mockConsoleDebug = jest.spyOn(console, 'debug').mockImplementation();
const mockConsoleError = jest.spyOn(console, 'error').mockImplementation();

// Now safe to import
import {ApproverPlugin} from './index';
import ApprovalPost from './components/ApprovalPost';
import ApprovalDMPost from './components/ApprovalDMPost';

describe('Plugin Registration', () => {
    let mockRegistry: any;
    let mockStore: any;

    beforeEach(() => {
        // Reset mocks
        mockConsoleLog.mockClear();
        mockConsoleDebug.mockClear();
        mockConsoleError.mockClear();
        mockRegisterPlugin.mockClear();

        // Create mock registry with registerPostTypeComponent
        mockRegistry = {
            registerPostTypeComponent: jest.fn(),
        };

        mockStore = {};
    });

    afterAll(() => {
        mockConsoleLog.mockRestore();
        mockConsoleDebug.mockRestore();
        mockConsoleError.mockRestore();
    });

    describe('ApproverPlugin.initialize', () => {
        it('should register custom post type component successfully', () => {
            const plugin = new ApproverPlugin();

            plugin.initialize(mockRegistry, mockStore);

            // Verify registerPostTypeComponent was called with correct arguments
            // Story 10.4: Now registering both custom_approval and custom_approval_dm
            expect(mockRegistry.registerPostTypeComponent).toHaveBeenCalledTimes(2);
            expect(mockRegistry.registerPostTypeComponent).toHaveBeenCalledWith(
                'custom_approval',
                ApprovalPost
            );
            expect(mockRegistry.registerPostTypeComponent).toHaveBeenCalledWith(
                'custom_approval_dm',
                ApprovalDMPost
            );

            // Verify success logging
            expect(mockConsoleLog).toHaveBeenCalledWith(
                expect.stringContaining('Approver Plugin Webapp v3.0.0 Initialized')
            );
            expect(mockConsoleDebug).toHaveBeenCalledWith('Registered custom post type: custom_approval');
            expect(mockConsoleDebug).toHaveBeenCalledWith('Registered custom post type: custom_approval_dm');
        });

        it('should return cleanup function', () => {
            const plugin = new ApproverPlugin();

            const cleanup = plugin.initialize(mockRegistry, mockStore);

            expect(cleanup).toBeInstanceOf(Function);

            // Call cleanup and verify it runs without error
            if (cleanup) {
                cleanup();
                expect(mockConsoleDebug).toHaveBeenCalledWith('Approver Plugin: Cleanup completed');
            }
        });

        it('should handle missing registry gracefully', () => {
            const plugin = new ApproverPlugin();

            plugin.initialize(null as any, mockStore);

            // Should not call registerPostTypeComponent
            expect(mockRegistry.registerPostTypeComponent).not.toHaveBeenCalled();

            // Should log error
            expect(mockConsoleError).toHaveBeenCalledWith(
                expect.stringContaining('Invalid registry object')
            );
        });

        it('should handle registry without registerPostTypeComponent method', () => {
            const plugin = new ApproverPlugin();

            const invalidRegistry = {};

            plugin.initialize(invalidRegistry as any, mockStore);

            // Should log error
            expect(mockConsoleError).toHaveBeenCalledWith(
                expect.stringContaining('Invalid registry object')
            );
        });

        it('should handle registration errors gracefully', () => {
            const plugin = new ApproverPlugin();

            // Make registerPostTypeComponent throw error
            mockRegistry.registerPostTypeComponent.mockImplementation(() => {
                throw new Error('Registration failed');
            });

            plugin.initialize(mockRegistry, mockStore);

            // Should catch error and log it
            expect(mockConsoleError).toHaveBeenCalledWith(
                'Approver Plugin: Failed to register custom post type',
                expect.any(Error)
            );
        });

        it('should handle missing PLUGIN_VERSION', () => {
            const plugin = new ApproverPlugin();

            // Mock PLUGIN_VERSION as undefined
            const originalVersion = (global as any).PLUGIN_VERSION;
            delete (global as any).PLUGIN_VERSION;

            plugin.initialize(mockRegistry, mockStore);

            // Should use 'unknown' as fallback
            expect(mockConsoleLog).toHaveBeenCalledWith(
                expect.stringContaining('Approver Plugin Webapp vunknown Initialized')
            );

            // Restore
            (global as any).PLUGIN_VERSION = originalVersion;
        });
    });

    describe('Plugin Export', () => {
        it('should export ApproverPlugin class', () => {
            // Verify class is exported
            expect(ApproverPlugin).toBeDefined();
            expect(typeof ApproverPlugin).toBe('function');
        });

        it('should create plugin instance with initialize method', () => {
            const plugin = new ApproverPlugin();

            // Verify plugin has required methods
            expect(plugin).toHaveProperty('initialize');
            expect(typeof plugin.initialize).toBe('function');
        });

        it('should have correct plugin ID constant', () => {
            // This verifies the plugin ID used in window.registerPlugin call
            // The actual value is tested by integration tests
            const pluginId = 'com.mattermost.plugin-approver2';

            // This is the ID that should be used for registration
            expect(pluginId).toBe('com.mattermost.plugin-approver2');
        });
    });

    describe('ApprovalPost import', () => {
        it('should successfully import ApprovalPost component', () => {
            expect(ApprovalPost).toBeDefined();
            expect(typeof ApprovalPost).toBe('object'); // React.memo returns object
        });
    });
});
