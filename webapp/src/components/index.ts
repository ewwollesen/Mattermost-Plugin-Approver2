/**
 * Approval Post UI Components Library
 *
 * Reusable components for building approval posts and notifications.
 * All components are memoized and theme-aware.
 */

// Story 9.6: ApprovalPost Base Component
export { default as ApprovalPost, ApprovalPostProps, ApprovalPostData } from './ApprovalPost';

// Story 9.5: UI Component Library
export { default as StatusBadge } from './StatusBadge';
export { default as UserMention } from './UserMention';
export { default as InfoRow } from './InfoRow';

// Story 9.4: Timestamp Component
export { default as Timestamp } from './Timestamp';

// Story 10.4: ApprovalDMPost Component for DM Notifications
export { default as ApprovalDMPost } from './ApprovalDMPost';

// Story 11.1: Modal Infrastructure Components
export { default as Modal } from './Modal';
export { default as ModalTriggerPost } from './ModalTriggerPost';

// Story 11.2: User Selector Component
export { default as UserSelector } from './UserSelector';
export type { UserSelectorProps, UserOption } from './UserSelector';

// Story 11.3: TextArea Component
export { default as TextArea } from './TextArea';
export type { TextAreaProps } from './TextArea';

// Story 11.3: Approval Request Modal Component
export { default as ApprovalRequestModal } from './ApprovalRequestModal';
export type { ApprovalRequestModalProps } from './ApprovalRequestModal';
