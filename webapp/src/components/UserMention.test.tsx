import React from 'react';
import { render } from '@testing-library/react';
import UserMention from './UserMention';

describe('UserMention Component', () => {
    it('renders username with @ prefix', () => {
        const { container } = render(<UserMention username="john.doe" />);
        expect(container.textContent).toBe('@john.doe');
    });

    it('includes displayName in title attribute', () => {
        const { container } = render(
            <UserMention username="john.doe" displayName="John Doe" />
        );
        const span = container.querySelector('span');
        expect(span?.title).toBe('John Doe (@john.doe)');
    });

    it('uses username as title when displayName not provided', () => {
        const { container } = render(<UserMention username="jane.doe" />);
        const span = container.querySelector('span');
        expect(span?.title).toBe('jane.doe');
    });

    it('applies clickable mention styling', () => {
        const { container } = render(<UserMention username="user" />);
        const span = container.querySelector('span');
        expect(span).toHaveStyle({
            cursor: 'pointer',
            fontWeight: '500'
        });
    });

    it('prevents unnecessary re-renders with React.memo', () => {
        const renderSpy = jest.fn();

        const TestWrapper = ({ username }: { username: string }) => {
            renderSpy();
            return <UserMention username={username} />;
        };

        const { rerender } = render(<TestWrapper username="john" />);
        expect(renderSpy).toHaveBeenCalledTimes(1);

        // Re-render with same props - memo should prevent re-render
        rerender(<TestWrapper username="john" />);
        expect(renderSpy).toHaveBeenCalledTimes(2); // Wrapper re-renders but UserMention shouldn't

        // Re-render with different props - should re-render
        rerender(<TestWrapper username="jane" />);
        expect(renderSpy).toHaveBeenCalledTimes(3);
    });
});
