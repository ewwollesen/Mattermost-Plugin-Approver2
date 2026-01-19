import React from 'react';
import {render} from '@testing-library/react';
import {Provider} from 'react-redux';
import configureStore from 'redux-mock-store';
import Timestamp from './Timestamp';

const mockStore = configureStore([]);

const createMockState = (overrides = {}) => ({
    entities: {
        users: {
            currentUserId: 'user1',
            profiles: {
                user1: {id: 'user1', username: 'testuser'}
            }
        },
        timezone: {},
        ...overrides
    }
});

describe('Timestamp Component', () => {
    const testTimestamp = 1705593000000; // Jan 18, 2024 10:30 AM UTC

    it('renders with default format', () => {
        const store = mockStore(createMockState());

        const {container} = render(
            <Provider store={store}>
                <Timestamp unixMillis={testTimestamp} />
            </Provider>
        );

        expect(container.textContent).toContain('Jan');
        expect(container.textContent).toContain('2024');
    });

    it('handles zero timestamp', () => {
        const store = mockStore(createMockState());
        const {container} = render(
            <Provider store={store}>
                <Timestamp unixMillis={0} />
            </Provider>
        );

        expect(container.textContent).toBe('Not yet decided');
    });

    it('handles null timestamp', () => {
        const store = mockStore(createMockState());
        const {container} = render(
            <Provider store={store}>
                {/* eslint-disable-next-line @typescript-eslint/no-explicit-any */}
                <Timestamp unixMillis={null as any} />
            </Provider>
        );

        expect(container.textContent).toBe('Not yet decided');
    });

    it('handles invalid timestamp', () => {
        const store = mockStore(createMockState());
        const {container} = render(
            <Provider store={store}>
                <Timestamp unixMillis={NaN} />
            </Provider>
        );

        expect(container.textContent).toBe('Invalid date');
    });

    it('displays relative time', () => {
        const store = mockStore(createMockState());
        const fiveMinutesAgo = Date.now() - (5 * 60 * 1000);

        const {container} = render(
            <Provider store={store}>
                <Timestamp unixMillis={fiveMinutesAgo} relative />
            </Provider>
        );

        expect(container.textContent).toContain('minutes ago');
    });

    it('uses custom format', () => {
        const store = mockStore(createMockState());

        const {container} = render(
            <Provider store={store}>
                <Timestamp unixMillis={testTimestamp} format="YYYY-MM-DD" />
            </Provider>
        );

        expect(container.textContent).toMatch(/2024-01-18/);
    });

    it('displays timezone in title attribute', () => {
        const store = mockStore(createMockState());

        const {container} = render(
            <Provider store={store}>
                <Timestamp unixMillis={testTimestamp} />
            </Provider>
        );

        const span = container.querySelector('span');
        expect(span?.title).toBeTruthy();
        expect(span?.title).toContain('2024');
        // Verify timezone abbreviation is present in tooltip
        expect(span?.title).toMatch(/[A-Z]{3,4}$/); // Matches PST, EST, GMT, etc.
    });

    it('handles negative timestamp', () => {
        const store = mockStore(createMockState());
        const {container} = render(
            <Provider store={store}>
                <Timestamp unixMillis={-1000} />
            </Provider>
        );

        expect(container.textContent).toBe('Invalid date');
    });

    it('converts timezone accurately for different timezones', () => {
        // Test with America/Los_Angeles timezone
        const storePST = mockStore({
            entities: {
                users: {
                    currentUserId: 'user1',
                    profiles: {
                        user1: {
                            id: 'user1',
                            username: 'testuser',
                            timezone: {
                                useAutomaticTimezone: 'true',
                                automaticTimezone: 'America/Los_Angeles',
                                manualTimezone: ''
                            }
                        }
                    }
                },
                timezone: {}
            }
        });

        const {container: containerPST} = render(
            <Provider store={storePST}>
                <Timestamp unixMillis={testTimestamp} />
            </Provider>
        );

        // Test with America/New_York timezone
        const storeEST = mockStore({
            entities: {
                users: {
                    currentUserId: 'user1',
                    profiles: {
                        user1: {
                            id: 'user1',
                            username: 'testuser',
                            timezone: {
                                useAutomaticTimezone: 'true',
                                automaticTimezone: 'America/New_York',
                                manualTimezone: ''
                            }
                        }
                    }
                },
                timezone: {}
            }
        });

        const {container: containerEST} = render(
            <Provider store={storeEST}>
                <Timestamp unixMillis={testTimestamp} />
            </Provider>
        );

        // PST and EST should display different times for same Unix timestamp (3 hour difference)
        expect(containerPST.textContent).not.toBe(containerEST.textContent);
    });

    it('respects useAutomaticTimezone vs manualTimezone preference', () => {
        // Test automatic timezone (useAutomaticTimezone: 'true' as string)
        const storeAuto = mockStore({
            entities: {
                users: {
                    currentUserId: 'user1',
                    profiles: {
                        user1: {
                            id: 'user1',
                            username: 'testuser',
                            timezone: {
                                useAutomaticTimezone: 'true',
                                automaticTimezone: 'America/Los_Angeles',
                                manualTimezone: 'America/New_York'
                            }
                        }
                    }
                },
                timezone: {}
            }
        });

        const {container: containerAuto} = render(
            <Provider store={storeAuto}>
                <Timestamp unixMillis={testTimestamp} />
            </Provider>
        );

        // Test manual timezone (useAutomaticTimezone: 'false' should use manualTimezone)
        const storeManual = mockStore({
            entities: {
                users: {
                    currentUserId: 'user1',
                    profiles: {
                        user1: {
                            id: 'user1',
                            username: 'testuser',
                            timezone: {
                                useAutomaticTimezone: 'false',
                                automaticTimezone: 'America/Los_Angeles',
                                manualTimezone: 'America/New_York'
                            }
                        }
                    }
                },
                timezone: {}
            }
        });

        const {container: containerManual} = render(
            <Provider store={storeManual}>
                <Timestamp unixMillis={testTimestamp} />
            </Provider>
        );

        // Automatic uses LA, manual uses NY - should differ (3 hour difference)
        expect(containerAuto.textContent).not.toBe(containerManual.textContent);
    });

    it('updates when timezone changes in Redux store', () => {
        const store = mockStore({
            entities: {
                users: {
                    currentUserId: 'user1',
                    profiles: {
                        user1: {
                            id: 'user1',
                            username: 'testuser',
                            timezone: {
                                useAutomaticTimezone: 'true',
                                automaticTimezone: 'America/Los_Angeles',
                                manualTimezone: ''
                            }
                        }
                    }
                },
                timezone: {}
            }
        });

        const {container, rerender} = render(
            <Provider store={store}>
                <Timestamp unixMillis={testTimestamp} />
            </Provider>
        );

        const initialText = container.textContent;

        // Change timezone in store (LA to NY = 3 hour difference)
        const updatedStore = mockStore({
            entities: {
                users: {
                    currentUserId: 'user1',
                    profiles: {
                        user1: {
                            id: 'user1',
                            username: 'testuser',
                            timezone: {
                                useAutomaticTimezone: 'true',
                                automaticTimezone: 'America/New_York',
                                manualTimezone: ''
                            }
                        }
                    }
                },
                timezone: {}
            }
        });

        rerender(
            <Provider store={updatedStore}>
                <Timestamp unixMillis={testTimestamp} />
            </Provider>
        );

        // Component should reflect new timezone (3 hours earlier in LA vs NY)
        expect(container.textContent).not.toBe(initialText);
    });
});
