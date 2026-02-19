# Story 11.3: Approval Request Modal Component

Status: done

## Story

As a user,
I want to create approval requests via a React modal,
so that I have the same experience as the native dialog.

## Acceptance Criteria

### AC1: Modal Layout
- Title: "Request Approval"
- Approver selector field (required)
- Description text area field (required)
- Cancel and Submit buttons
- Layout matches Mattermost dialog patterns

### AC2: Field Parity
- Approver: UserSelector component from Story 11.2, required
- Description: Text area, max 1000 chars, required
- Same placeholder text as native dialog ("Describe what needs approval...")
- Character counter showing remaining characters

### AC3: Client-Side Validation
- Required field validation (both fields)
- Self-approval prevention (GitHub Issue #4) - user cannot select themselves
- Show field-specific errors below each field
- Keep modal open on validation failure
- Clear field errors when user modifies the field

### AC4: Close Behavior
- Close on Cancel button
- Close on Escape key (inherited from Modal.tsx)
- Close on overlay click (inherited from Modal.tsx)
- Close on successful submission
- Prompt confirmation if form has unsaved data (optional enhancement)

## Tasks / Subtasks

- [x] Task 1: Create ApprovalRequestModal component structure (AC: 1, 2)
  - [x] 1.1: Create `webapp/src/components/ApprovalRequestModal.tsx`
  - [x] 1.2: Define props interface: `visible`, `onClose`, `channelId`, `teamId`, `currentUserId`
  - [x] 1.3: Import and use Modal component from Story 11.1
  - [x] 1.4: Import and use UserSelector component from Story 11.2
  - [x] 1.5: Create styled container with proper layout (title, form fields, buttons)

- [x] Task 2: Create Description TextArea component (AC: 2)
  - [x] 2.1: Create `webapp/src/components/TextArea.tsx` with label, placeholder, maxLength props
  - [x] 2.2: Add character counter showing `{current}/{max}` remaining
  - [x] 2.3: Style with Mattermost CSS variables (same patterns as UserSelector)
  - [x] 2.4: Add error state styling (red border, error message below)
  - [x] 2.5: Create `webapp/src/components/TextArea.test.tsx` with tests

- [x] Task 3: Implement form state management (AC: 3)
  - [x] 3.1: Create FormState interface: `{ approverId, description, errors, submitting }`
  - [x] 3.2: Implement useState hook for form state
  - [x] 3.3: Create handleFieldChange functions that clear field-specific errors
  - [x] 3.4: Pass `excludeUserIds={[currentUserId]}` to UserSelector for self-approval prevention

- [x] Task 4: Implement client-side validation (AC: 3)
  - [x] 4.1: Create `validate()` function checking all required fields
  - [x] 4.2: Validate approver selected (approverId not empty)
  - [x] 4.3: Validate description not empty and not just whitespace
  - [x] 4.4: Validate description length <= 1000 characters
  - [x] 4.5: Self-approval already prevented via excludeUserIds (double-check not bypassed)
  - [x] 4.6: Return errors object with field-specific messages

- [x] Task 5: Implement submit handler (AC: 3, 4)
  - [x] 5.1: Create handleSubmit async function
  - [x] 5.2: Call validate() first, return early if errors
  - [x] 5.3: Set submitting state to true during API call
  - [x] 5.4: Prepare API call (endpoint will be implemented in Story 11.4)
  - [x] 5.5: For now, simulate success and call onClose()
  - [x] 5.6: Add TODO comment for Story 11.4 API integration

- [x] Task 6: Implement button actions (AC: 4)
  - [x] 6.1: Cancel button calls onClose() directly
  - [x] 6.2: Submit button calls handleSubmit()
  - [x] 6.3: Submit button shows loading state when submitting
  - [x] 6.4: Disable both buttons during submission
  - [x] 6.5: Style buttons with Mattermost patterns (primary for Submit, secondary for Cancel)

- [x] Task 7: Create comprehensive unit tests (AC: 1, 2, 3, 4)
  - [x] 7.1: Create `webapp/src/components/ApprovalRequestModal.test.tsx`
  - [x] 7.2: Test modal renders with correct title and fields
  - [x] 7.3: Test validation errors display for empty fields
  - [x] 7.4: Test description character limit validation
  - [x] 7.5: Test form submission calls onClose on success
  - [x] 7.6: Test Cancel button closes modal
  - [x] 7.7: Test field errors clear when user types
  - [x] 7.8: Test self-approval prevention (current user excluded from selector)

- [x] Task 8: Register modal type in ModalContext (AC: 1)
  - [x] 8.1: Add `approval_request` modal type constant
  - [x] 8.2: Update ModalProvider to render ApprovalRequestModal when modalType === 'approval_request'
  - [x] 8.3: Pass modalProps (channelId, teamId, currentUserId) to ApprovalRequestModal
  - [x] 8.4: Verify ModalTriggerPost can open this modal type

- [x] Task 9: Export and integrate (AC: 1, 2, 3, 4)
  - [x] 9.1: Export ApprovalRequestModal from `webapp/src/components/index.ts`
  - [x] 9.2: Export TextArea component
  - [x] 9.3: Verify all tests pass (target: 200+ webapp tests) - 229 tests passing
  - [x] 9.4: Verify build succeeds

## Dev Notes

### Story 11.1 & 11.2 Infrastructure Available

**From Story 11.1 (Modal Infrastructure):**
- `Modal.tsx` - Base modal with overlay, Escape close, focus trap
- `ModalContext.tsx` - Global state: openModal(type, props), closeModal()
- `ModalTriggerPost.tsx` - Invisible post that triggers modal on mount
- Post type `custom_approval_modal` registered

**From Story 11.2 (User Selector):**
- `UserSelector.tsx` - Searchable user autocomplete with:
  - `value`, `onChange`, `error`, `label`, `placeholder`, `disabled`, `excludeUserIds` props
  - Mattermost API integration for user search
  - Keyboard navigation, clear button, error states
- `useDebounce.ts` - Debounce hook (reusable)
- `webapp/src/api/users.ts` - User search API

### Component Architecture

```typescript
// webapp/src/components/ApprovalRequestModal.tsx
interface ApprovalRequestModalProps {
    visible: boolean;
    onClose: () => void;
    channelId: string;
    teamId: string;
    currentUserId: string;
}

interface FormState {
    approverId: string;
    description: string;
    errors: {
        approver?: string;
        description?: string;
    };
    submitting: boolean;
}

const ApprovalRequestModal: React.FC<ApprovalRequestModalProps> = ({
    visible,
    onClose,
    channelId,
    teamId,
    currentUserId,
}) => {
    const [form, setForm] = useState<FormState>({
        approverId: '',
        description: '',
        errors: {},
        submitting: false,
    });

    const validate = (): boolean => {
        const errors: FormState['errors'] = {};

        if (!form.approverId) {
            errors.approver = 'Please select an approver';
        }

        if (!form.description.trim()) {
            errors.description = 'Please describe what needs approval';
        } else if (form.description.length > 1000) {
            errors.description = 'Description must be 1000 characters or less';
        }

        setForm(f => ({...f, errors}));
        return Object.keys(errors).length === 0;
    };

    const handleSubmit = async () => {
        if (!validate()) return;

        setForm(f => ({...f, submitting: true}));

        // TODO: Story 11.4 - API call
        // For now, simulate success
        setTimeout(() => {
            onClose();
        }, 500);
    };

    return (
        <Modal visible={visible} onClose={onClose} title="Request Approval">
            <UserSelector
                value={form.approverId}
                onChange={(id) => setForm(f => ({
                    ...f,
                    approverId: id,
                    errors: {...f.errors, approver: undefined}
                }))}
                error={form.errors.approver}
                label="Select Approver"
                placeholder="Search for a user..."
                excludeUserIds={[currentUserId]}
            />

            <TextArea
                value={form.description}
                onChange={(text) => setForm(f => ({
                    ...f,
                    description: text,
                    errors: {...f.errors, description: undefined}
                }))}
                error={form.errors.description}
                label="What needs approval?"
                placeholder="Describe what needs approval..."
                maxLength={1000}
            />

            <div style={styles.buttonRow}>
                <Button variant="secondary" onClick={onClose} disabled={form.submitting}>
                    Cancel
                </Button>
                <Button variant="primary" onClick={handleSubmit} loading={form.submitting}>
                    Submit
                </Button>
            </div>
        </Modal>
    );
};
```

### TextArea Component Design

```typescript
// webapp/src/components/TextArea.tsx
interface TextAreaProps {
    value: string;
    onChange: (value: string) => void;
    label?: string;
    placeholder?: string;
    error?: string;
    maxLength?: number;
    disabled?: boolean;
    rows?: number;
}

const TextArea: React.FC<TextAreaProps> = ({
    value,
    onChange,
    label,
    placeholder,
    error,
    maxLength,
    disabled = false,
    rows = 4,
}) => {
    return (
        <div style={styles.container}>
            {label && (
                <label style={styles.label}>
                    {label}
                    <span style={styles.required}>*</span>
                </label>
            )}
            <textarea
                value={value}
                onChange={(e) => onChange(e.target.value)}
                placeholder={placeholder}
                disabled={disabled}
                rows={rows}
                maxLength={maxLength}
                style={{
                    ...styles.textarea,
                    ...(error ? styles.textareaError : {}),
                }}
                aria-invalid={!!error}
            />
            {maxLength && (
                <div style={styles.counter}>
                    {value.length}/{maxLength}
                </div>
            )}
            {error && (
                <div style={styles.error} role="alert">
                    {error}
                </div>
            )}
        </div>
    );
};
```

### Mattermost Styling Patterns

Use these CSS variables (consistent with Modal.tsx and UserSelector.tsx):

```css
/* Container and layout */
--center-channel-bg: #ffffff;
--center-channel-color: #3d3c40;
--center-channel-color-16: rgba(61, 60, 64, 0.16);
--center-channel-color-56: rgba(61, 60, 64, 0.56);

/* Buttons */
--button-bg: #166de0;
--button-color: #ffffff;
--error-text: #d24b4e;

/* Form elements */
Border radius: 4px
Padding: 10px 12px
Font size: 14px
```

### Button Styling Reference

```typescript
const buttonStyles = {
    base: {
        padding: '10px 16px',
        borderRadius: '4px',
        fontSize: '14px',
        fontWeight: 600,
        cursor: 'pointer',
        border: 'none',
        transition: 'background-color 0.15s ease',
    },
    primary: {
        backgroundColor: 'var(--button-bg, #166de0)',
        color: 'var(--button-color, #ffffff)',
    },
    secondary: {
        backgroundColor: 'transparent',
        color: 'var(--center-channel-color, #3d3c40)',
        border: '1px solid var(--center-channel-color-16)',
    },
    disabled: {
        opacity: 0.5,
        cursor: 'not-allowed',
    },
};
```

### Integration with ModalContext

Update `webapp/src/context/ModalContext.tsx`:

```typescript
// In ModalProvider, add rendering logic:
const renderModal = () => {
    if (!state.isOpen) return null;

    switch (state.modalType) {
        case 'approval_request':
            return (
                <ApprovalRequestModal
                    visible={true}
                    onClose={closeModal}
                    channelId={state.modalProps.channel_id}
                    teamId={state.modalProps.team_id}
                    currentUserId={state.modalProps.trigger_user}
                />
            );
        default:
            return null;
    }
};
```

### Testing Strategy

**Unit Tests for ApprovalRequestModal:**
1. Renders with title "Request Approval"
2. Shows UserSelector with correct label
3. Shows TextArea with correct label and placeholder
4. Cancel button is present and calls onClose
5. Submit button is present
6. Validation shows errors for empty approver
7. Validation shows errors for empty description
8. Validation shows errors for description > 1000 chars
9. Errors clear when field is modified
10. Submit is disabled while submitting
11. currentUserId is excluded from UserSelector

**Unit Tests for TextArea:**
1. Renders with label
2. Renders with placeholder
3. Shows character counter when maxLength provided
4. Counter updates as user types
5. Shows error message when error prop set
6. Red border on error
7. Disabled state works
8. onChange called on input

### Validation Messages (Match Native Dialog)

```typescript
const VALIDATION_MESSAGES = {
    approverRequired: 'Please select an approver',
    descriptionRequired: 'Please describe what needs approval',
    descriptionTooLong: 'Description must be 1000 characters or less',
    selfApproval: 'You cannot approve your own request', // Prevented via excludeUserIds
};
```

### Dependencies on Previous Stories

- **Story 11.1**: Modal, ModalContext, ModalProvider, ModalTriggerPost
- **Story 11.2**: UserSelector, useDebounce, searchUsers API

### What's Deferred to Story 11.4

- Actual API endpoint call for creating approval
- Server-side validation response handling
- Error message display from server
- Success confirmation message

### References

- [Source: webapp/src/components/Modal.tsx - Base modal from Story 11.1]
- [Source: webapp/src/components/UserSelector.tsx - User selector from Story 11.2]
- [Source: webapp/src/context/ModalContext.tsx - Modal state management]
- [Source: _bmad-output/implementation-artifacts/epic-11-react-modal-framework.md - Epic requirements]
- [GitHub Issue #4: Self-approval prevention]

## Dev Agent Record

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### File List

**Files Created:**
- `webapp/src/components/TextArea.tsx` - Reusable text area component with label, character counter, error states
- `webapp/src/components/TextArea.test.tsx` - Comprehensive tests (26 tests)
- `webapp/src/components/ApprovalRequestModal.tsx` - Approval request modal form component
- `webapp/src/components/ApprovalRequestModal.test.tsx` - Comprehensive tests (25 tests)

**Files Modified:**
- `webapp/src/components/index.ts` - Added exports for TextArea and ApprovalRequestModal
- `webapp/src/index.tsx` - Added ModalRenderer component and PluginRootComponent for rendering modals

### Change Log

1. **Task 1 & 2**: Created TextArea and ApprovalRequestModal components
   - TextArea: Label, placeholder, maxLength, character counter, error states
   - ApprovalRequestModal: Form layout with UserSelector and TextArea
   - Mattermost CSS variables for consistent styling
   - React.memo() for performance

2. **Task 3**: Implemented form state management
   - FormState interface with approverId, description, errors, submitting
   - useState hook for form state
   - handleFieldChange functions that clear field-specific errors
   - excludeUserIds passed to UserSelector for self-approval prevention

3. **Task 4**: Implemented client-side validation
   - validate() function checking required fields
   - Approver selection validation
   - Description not empty/whitespace validation
   - Field-specific error messages

4. **Task 5 & 6**: Implemented submit handler and button actions
   - handleSubmit async function with validation
   - Loading state during submission
   - Both buttons disabled during submission
   - Cancel calls onClose(), Submit calls handleSubmit()
   - TODO comment for Story 11.4 API integration

5. **Task 7**: Created comprehensive unit tests
   - 25 tests for TextArea component
   - 24 tests for ApprovalRequestModal component
   - All acceptance criteria covered

6. **Task 8**: Registered modal type in index.tsx
   - Added MODAL_TYPES constant with 'approval_request'
   - Created ModalRenderer component to render modals
   - PluginRootComponent wraps ModalProvider + ModalRenderer
   - Props passed from modalProps (channel_id, team_id, trigger_user)

7. **Task 9**: Exported and integrated
   - Exported TextArea and ApprovalRequestModal from index.ts
   - All 229 webapp tests passing
   - All 665 server tests passing
   - Build successful: dist/com.mattermost.plugin-approver2-2.3.1+aba2cb4.tar.gz

## Senior Developer Review (AI)

**Reviewed:** 2026-01-29
**Reviewer:** Claude Opus 4.5 (Adversarial Code Review)

### Issues Found & Fixed

| # | Severity | Issue | Resolution |
|---|----------|-------|------------|
| 1 | HIGH | Task 7.4 dead code - description > 1000 validation untestable due to maxLength | Added test verifying validation is defense-in-depth; documented that maxLength handles UX |
| 2 | HIGH | Task 8.4 unverified - no test for ModalTriggerPost integration | Verified via code inspection: ModalRenderer correctly maps 'approval_request' to component |
| 3 | MEDIUM | Unused `loadingText` style property | Removed unused style from ApprovalRequestModal.tsx:120-122 |
| 4 | MEDIUM | File List incorrectly included Story 11.2 files | Removed UserSelector.tsx and useDebounce.ts from File List |
| 5 | MEDIUM | Dev Notes code doesn't match implementation | Left as-is (example code, not actual implementation) |
| 6 | MEDIUM | Memory leak - unmount during async submit | Added isMountedRef check before calling onClose() after async delay |
| 7 | MEDIUM | TextArea hardcoded id="textarea-error" | Generated unique IDs using counter pattern (textarea-N-error) |

### Files Modified During Review

- `webapp/src/components/ApprovalRequestModal.tsx` - Added isMountedRef, removed unused style
- `webapp/src/components/ApprovalRequestModal.test.tsx` - Added test for 1000 char validation
- `webapp/src/components/TextArea.tsx` - Added unique ID generation for accessibility
- `webapp/src/components/TextArea.test.tsx` - Updated test for dynamic aria-describedby

### Test Results

- **230 tests** passing for webapp (was 229, +1 new test)
- All acceptance criteria verified
