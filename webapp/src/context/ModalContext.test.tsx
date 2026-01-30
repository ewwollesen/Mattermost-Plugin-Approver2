import React from 'react';
import {render, screen, act} from '@testing-library/react';
import {ModalProvider, useModal, dispatchModalOpen, dispatchModalClose} from './ModalContext';

// Test component to access modal context
const TestConsumer: React.FC = () => {
    const {state, openModal, closeModal} = useModal();

    return (
        <div>
            <span data-testid="is-open">{state.isOpen.toString()}</span>
            <span data-testid="modal-type">{state.modalType || 'null'}</span>
            <span data-testid="modal-props">{JSON.stringify(state.modalProps)}</span>
            <button data-testid="open-btn" onClick={() => openModal('test-modal', {foo: 'bar'})}>
                Open
            </button>
            <button data-testid="close-btn" onClick={closeModal}>
                Close
            </button>
        </div>
    );
};

describe('ModalContext (AC3)', () => {
    describe('Initial state', () => {
        it('starts with modal closed', () => {
            render(
                <ModalProvider>
                    <TestConsumer />
                </ModalProvider>
            );

            expect(screen.getByTestId('is-open').textContent).toBe('false');
            expect(screen.getByTestId('modal-type').textContent).toBe('null');
            expect(screen.getByTestId('modal-props').textContent).toBe('{}');
        });
    });

    describe('openModal action', () => {
        it('opens modal with type and props', () => {
            render(
                <ModalProvider>
                    <TestConsumer />
                </ModalProvider>
            );

            act(() => {
                screen.getByTestId('open-btn').click();
            });

            expect(screen.getByTestId('is-open').textContent).toBe('true');
            expect(screen.getByTestId('modal-type').textContent).toBe('test-modal');
            expect(screen.getByTestId('modal-props').textContent).toBe('{"foo":"bar"}');
        });

        it('opens modal with type only (no props)', () => {
            const TestConsumerTypeOnly: React.FC = () => {
                const {state, openModal} = useModal();
                return (
                    <div>
                        <span data-testid="is-open">{state.isOpen.toString()}</span>
                        <span data-testid="modal-type">{state.modalType || 'null'}</span>
                        <button data-testid="open-btn" onClick={() => openModal('simple-modal')}>
                            Open
                        </button>
                    </div>
                );
            };

            render(
                <ModalProvider>
                    <TestConsumerTypeOnly />
                </ModalProvider>
            );

            act(() => {
                screen.getByTestId('open-btn').click();
            });

            expect(screen.getByTestId('is-open').textContent).toBe('true');
            expect(screen.getByTestId('modal-type').textContent).toBe('simple-modal');
        });
    });

    describe('closeModal action', () => {
        it('closes modal and resets state', () => {
            render(
                <ModalProvider>
                    <TestConsumer />
                </ModalProvider>
            );

            // Open modal first
            act(() => {
                screen.getByTestId('open-btn').click();
            });

            expect(screen.getByTestId('is-open').textContent).toBe('true');

            // Close modal
            act(() => {
                screen.getByTestId('close-btn').click();
            });

            expect(screen.getByTestId('is-open').textContent).toBe('false');
            expect(screen.getByTestId('modal-type').textContent).toBe('null');
            expect(screen.getByTestId('modal-props').textContent).toBe('{}');
        });
    });

    describe('Multiple modal types support (AC3)', () => {
        it('can switch between different modal types', () => {
            const TestConsumerMultiModal: React.FC = () => {
                const {state, openModal, closeModal} = useModal();
                return (
                    <div>
                        <span data-testid="modal-type">{state.modalType || 'null'}</span>
                        <button data-testid="open-modal-a" onClick={() => openModal('modal-a')}>
                            Open A
                        </button>
                        <button data-testid="open-modal-b" onClick={() => openModal('modal-b')}>
                            Open B
                        </button>
                        <button data-testid="close-btn" onClick={closeModal}>
                            Close
                        </button>
                    </div>
                );
            };

            render(
                <ModalProvider>
                    <TestConsumerMultiModal />
                </ModalProvider>
            );

            // Open modal A
            act(() => {
                screen.getByTestId('open-modal-a').click();
            });
            expect(screen.getByTestId('modal-type').textContent).toBe('modal-a');

            // Switch to modal B (without closing first)
            act(() => {
                screen.getByTestId('open-modal-b').click();
            });
            expect(screen.getByTestId('modal-type').textContent).toBe('modal-b');

            // Close
            act(() => {
                screen.getByTestId('close-btn').click();
            });
            expect(screen.getByTestId('modal-type').textContent).toBe('null');
        });
    });

    describe('useModal hook outside provider', () => {
        // Suppress console.error for this test
        const originalError = console.error;
        beforeEach(() => {
            console.error = jest.fn();
        });
        afterEach(() => {
            console.error = originalError;
        });

        it('throws error when used outside ModalProvider', () => {
            const TestComponentOutsideProvider: React.FC = () => {
                useModal();
                return <div>Should not render</div>;
            };

            expect(() => {
                render(<TestComponentOutsideProvider />);
            }).toThrow('useModal must be used within a ModalProvider');
        });
    });

    describe('Global event handling (Issue 2 fix)', () => {
        it('opens modal via dispatchModalOpen global event', () => {
            render(
                <ModalProvider>
                    <TestConsumer />
                </ModalProvider>
            );

            expect(screen.getByTestId('is-open').textContent).toBe('false');

            // Dispatch global event
            act(() => {
                dispatchModalOpen('global-modal', {data: 'test'});
            });

            expect(screen.getByTestId('is-open').textContent).toBe('true');
            expect(screen.getByTestId('modal-type').textContent).toBe('global-modal');
            expect(screen.getByTestId('modal-props').textContent).toBe('{"data":"test"}');
        });

        it('closes modal via dispatchModalClose global event', () => {
            render(
                <ModalProvider>
                    <TestConsumer />
                </ModalProvider>
            );

            // First open the modal
            act(() => {
                dispatchModalOpen('test-modal');
            });

            expect(screen.getByTestId('is-open').textContent).toBe('true');

            // Dispatch global close event
            act(() => {
                dispatchModalClose();
            });

            expect(screen.getByTestId('is-open').textContent).toBe('false');
            expect(screen.getByTestId('modal-type').textContent).toBe('null');
        });

        it('handles dispatchModalOpen without props', () => {
            render(
                <ModalProvider>
                    <TestConsumer />
                </ModalProvider>
            );

            act(() => {
                dispatchModalOpen('simple-modal');
            });

            expect(screen.getByTestId('is-open').textContent).toBe('true');
            expect(screen.getByTestId('modal-type').textContent).toBe('simple-modal');
            expect(screen.getByTestId('modal-props').textContent).toBe('{}');
        });

        it('cleans up global event listeners on unmount', () => {
            const {unmount} = render(
                <ModalProvider>
                    <TestConsumer />
                </ModalProvider>
            );

            // Unmount the provider
            unmount();

            // Dispatching event after unmount should not cause errors
            // (listeners should be removed)
            expect(() => {
                act(() => {
                    dispatchModalOpen('orphan-modal');
                });
            }).not.toThrow();
        });
    });
});
