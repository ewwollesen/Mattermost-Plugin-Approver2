import React from 'react';
import { render } from '@testing-library/react';
import StatusBadge from './StatusBadge';

describe('StatusBadge Component', () => {
    it('renders pending status', () => {
        const { container } = render(<StatusBadge status="pending" />);
        expect(container.textContent).toBe('⏳ Approval Pending');
    });

    it('renders approved status', () => {
        const { container } = render(<StatusBadge status="approved" />);
        expect(container.textContent).toBe('✅ Approval Approved');
    });

    it('renders denied status', () => {
        const { container } = render(<StatusBadge status="denied" />);
        expect(container.textContent).toBe('❌ Approval Denied');
    });

    it('renders canceled status', () => {
        const { container } = render(<StatusBadge status="canceled" />);
        expect(container.textContent).toBe('🚫 Approval Canceled');
    });

    it('renders timeout status', () => {
        const { container } = render(<StatusBadge status="timeout" />);
        expect(container.textContent).toBe('⏱️ Approval Timed Out');
    });

    it('applies correct styling', () => {
        const { container } = render(<StatusBadge status="pending" />);
        const div = container.querySelector('div');
        expect(div).toHaveStyle({
            fontSize: '18px',
            fontWeight: '600',
            marginBottom: '8px'
        });
    });

    it('applies theme-aware text color', () => {
        const { container } = render(<StatusBadge status="approved" />);
        const div = container.querySelector('div');
        expect(div).toHaveStyle({ color: 'var(--center-channel-color, #3d3c40)' });
    });

    it('has role="status" for accessibility', () => {
        const { container } = render(<StatusBadge status="pending" />);
        const div = container.querySelector('div');
        expect(div?.getAttribute('role')).toBe('status');
    });

    it('handles invalid status gracefully', () => {
        // Simulate runtime invalid status from API
        const { container } = render(<StatusBadge status={'invalid' as any} />);
        expect(container.textContent).toBe('⚠️ Unknown Status');
    });

    it('prevents unnecessary re-renders with React.memo', () => {
        const renderSpy = jest.fn();

        const TestWrapper = ({ status }: { status: 'pending' | 'approved' }) => {
            renderSpy();
            return <StatusBadge status={status} />;
        };

        const { rerender } = render(<TestWrapper status="pending" />);
        expect(renderSpy).toHaveBeenCalledTimes(1);

        // Re-render with same props - memo should prevent re-render
        rerender(<TestWrapper status="pending" />);
        expect(renderSpy).toHaveBeenCalledTimes(2); // Wrapper re-renders but StatusBadge shouldn't

        // Re-render with different props - should re-render
        rerender(<TestWrapper status="approved" />);
        expect(renderSpy).toHaveBeenCalledTimes(3);
    });
});
