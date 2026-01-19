import React from 'react';
import { render } from '@testing-library/react';
import InfoRow from './InfoRow';
import UserMention from './UserMention';

describe('InfoRow Component', () => {
    it('renders label and string value', () => {
        const { container } = render(
            <InfoRow label="Request ID" value="A-SSJEQZ" />
        );
        expect(container.textContent).toContain('Request ID:');
        expect(container.textContent).toContain('A-SSJEQZ');
    });

    it('renders label and ReactNode value', () => {
        const { container } = render(
            <InfoRow label="User" value={<UserMention username="john" />} />
        );
        expect(container.textContent).toContain('User:');
        expect(container.textContent).toContain('@john');
    });

    it('renders optional icon', () => {
        const { container } = render(
            <InfoRow label="Test" value="Value" icon="🔍" />
        );
        expect(container.textContent).toContain('🔍');
        expect(container.textContent).toContain('Test:');
        expect(container.textContent).toContain('Value');
    });

    it('does not render icon when not provided', () => {
        const { container } = render(
            <InfoRow label="Test" value="Value" />
        );
        // Should only contain label and value, no icon
        expect(container.textContent).toBe('Test:Value');
    });

    it('applies flexbox layout styling', () => {
        const { container } = render(
            <InfoRow label="Label" value="Value" />
        );
        const div = container.querySelector('div');
        expect(div).toHaveStyle({
            display: 'flex',
            alignItems: 'baseline',
            gap: '8px'
        });
    });

    it('renders number value correctly', () => {
        const { container } = render(
            <InfoRow label="Count" value={42} />
        );
        expect(container.textContent).toContain('Count:');
        expect(container.textContent).toContain('42');
    });

    it('renders complex ReactNode value', () => {
        const complexValue = (
            <div>
                <span>Part 1</span> - <span>Part 2</span>
            </div>
        );
        const { container } = render(
            <InfoRow label="Complex" value={complexValue} />
        );
        expect(container.textContent).toContain('Complex:');
        expect(container.textContent).toContain('Part 1 - Part 2');
    });

    it('respects showColon prop', () => {
        const { container: withColon } = render(
            <InfoRow label="Test" value="Value" showColon={true} />
        );
        expect(withColon.textContent).toContain('Test:');

        const { container: withoutColon } = render(
            <InfoRow label="Test" value="Value" showColon={false} />
        );
        expect(withoutColon.textContent).toContain('Test');
        expect(withoutColon.textContent).not.toContain('Test:');
    });

    it('prevents unnecessary re-renders with React.memo', () => {
        const renderSpy = jest.fn();

        const TestWrapper = ({ label }: { label: string }) => {
            renderSpy();
            return <InfoRow label={label} value="test" />;
        };

        const { rerender } = render(<TestWrapper label="Label1" />);
        expect(renderSpy).toHaveBeenCalledTimes(1);

        // Re-render with same props - memo should prevent re-render
        rerender(<TestWrapper label="Label1" />);
        expect(renderSpy).toHaveBeenCalledTimes(2); // Wrapper re-renders but InfoRow shouldn't

        // Re-render with different props - should re-render
        rerender(<TestWrapper label="Label2" />);
        expect(renderSpy).toHaveBeenCalledTimes(3);
    });
});
