import React, {createContext, useContext, useState, useCallback, useMemo, useEffect} from 'react';

/**
 * Issue 2 fix: Global modal event system
 * Provides fallback when ModalProvider context isn't accessible
 * (e.g., when custom post types are rendered outside the React tree)
 */
const MODAL_OPEN_EVENT = 'approver-modal-open';
const MODAL_CLOSE_EVENT = 'approver-modal-close';

interface ModalEventDetail {
    type: string;
    props: Record<string, any>;
}

/**
 * Dispatch a global event to open a modal
 * Used as fallback when React context isn't available
 */
export const dispatchModalOpen = (type: string, props: Record<string, any> = {}): void => {
    const event = new CustomEvent<ModalEventDetail>(MODAL_OPEN_EVENT, {
        detail: {type, props},
    });
    window.dispatchEvent(event);
};

/**
 * Dispatch a global event to close the modal
 */
export const dispatchModalClose = (): void => {
    window.dispatchEvent(new Event(MODAL_CLOSE_EVENT));
};

/**
 * Modal State Interface
 * Story 11.1 - AC3: Global Modal State
 */
export interface ModalState {
    /** Whether a modal is currently open */
    isOpen: boolean;
    /** The type of modal currently open (null if none) */
    modalType: string | null;
    /** Props to pass to the modal component */
    modalProps: Record<string, any>;
}

/**
 * Modal Context Value Interface
 * Provides state and actions for modal management
 */
export interface ModalContextValue {
    /** Current modal state */
    state: ModalState;
    /** Open a modal with the specified type and optional props */
    openModal: (type: string, props?: Record<string, any>) => void;
    /** Close the currently open modal */
    closeModal: () => void;
}

/**
 * Initial modal state - modal closed with no type or props
 */
const initialState: ModalState = {
    isOpen: false,
    modalType: null,
    modalProps: {},
};

/**
 * Modal Context
 * Used internally - consumers should use useModal() hook
 */
const ModalContext = createContext<ModalContextValue | undefined>(undefined);

/**
 * Modal Provider Props
 */
interface ModalProviderProps {
    children: React.ReactNode;
}

/**
 * Modal Provider Component
 *
 * Provides modal state management to the component tree.
 * Wrap your app or plugin root with this provider to enable modal functionality.
 *
 * @example
 * <ModalProvider>
 *   <App />
 * </ModalProvider>
 */
export const ModalProvider: React.FC<ModalProviderProps> = ({children}) => {
    const [state, setState] = useState<ModalState>(initialState);

    /**
     * Open a modal with the specified type and optional props
     * AC3: Support multiple modal types
     */
    const openModal = useCallback((type: string, props: Record<string, any> = {}) => {
        setState({
            isOpen: true,
            modalType: type,
            modalProps: props,
        });
    }, []);

    /**
     * Close the currently open modal and reset state
     */
    const closeModal = useCallback(() => {
        setState(initialState);
    }, []);

    /**
     * Issue 2 fix: Listen for global modal events
     * This enables components outside the React context tree to open modals
     */
    useEffect(() => {
        const handleGlobalOpen = (event: Event) => {
            const customEvent = event as CustomEvent<ModalEventDetail>;
            if (customEvent.detail) {
                openModal(customEvent.detail.type, customEvent.detail.props);
            }
        };

        const handleGlobalClose = () => {
            closeModal();
        };

        window.addEventListener(MODAL_OPEN_EVENT, handleGlobalOpen);
        window.addEventListener(MODAL_CLOSE_EVENT, handleGlobalClose);

        return () => {
            window.removeEventListener(MODAL_OPEN_EVENT, handleGlobalOpen);
            window.removeEventListener(MODAL_CLOSE_EVENT, handleGlobalClose);
        };
    }, [openModal, closeModal]);

    // Memoize context value to prevent unnecessary re-renders
    const contextValue = useMemo<ModalContextValue>(() => ({
        state,
        openModal,
        closeModal,
    }), [state, openModal, closeModal]);

    return (
        <ModalContext.Provider value={contextValue}>
            {children}
        </ModalContext.Provider>
    );
};

/**
 * useModal Hook
 *
 * Access modal state and actions from any component within ModalProvider.
 *
 * @throws Error if used outside of ModalProvider
 *
 * @example
 * const {state, openModal, closeModal} = useModal();
 *
 * // Open a modal
 * openModal('approval_request', {channelId: '123'});
 *
 * // Close the modal
 * closeModal();
 *
 * // Check if modal is open
 * if (state.isOpen && state.modalType === 'approval_request') {
 *   // Render approval request modal
 * }
 */
export const useModal = (): ModalContextValue => {
    const context = useContext(ModalContext);

    if (context === undefined) {
        throw new Error('useModal must be used within a ModalProvider');
    }

    return context;
};

export default ModalContext;
