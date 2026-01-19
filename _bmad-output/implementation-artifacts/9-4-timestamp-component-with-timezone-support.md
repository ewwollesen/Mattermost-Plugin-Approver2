# Story 9.4: Timestamp Component with Timezone Support

Status: done

## Story

As an approval post component,
I want a reusable Timestamp component that displays times in user's timezone,
so that all approval timestamps respect user preferences.

## Acceptance Criteria

**AC1: Timestamp Component**
- Create `webapp/src/components/Timestamp.tsx`
- Props interface:
```typescript
interface TimestampProps {
    unixMillis: number;       // Unix timestamp in milliseconds
    format?: string;          // moment format string (default: 'lll')
    relative?: boolean;       // Show relative time ("5 minutes ago")
}
```

**AC2: Timezone Conversion**
- Use `getCurrentTimezone()` from `mattermost-redux/selectors/entities/timezone`
- Convert Unix timestamp to user's timezone using moment-timezone
- Handle null/undefined timezone (fallback to browser timezone)
- Update when user changes timezone setting (React to selector changes)

**AC3: Format Options**
- Default format: "Jan 18, 2026 10:30 AM" (moment 'lll')
- Relative format: "5 minutes ago" (moment.fromNow())
- Custom format: Accept moment format strings
- Display timezone abbreviation on hover (tooltip)

**AC4: Edge Cases**
- Handle 0 or null timestamps (display "Not yet decided" or similar)
- Handle invalid timestamps gracefully
- Handle missing timezone data
- Performance: memo component to prevent unnecessary re-renders

**AC5: Unit Tests**
- Test timezone conversion accuracy
- Test format options
- Test edge cases (0, null, invalid)
- Test relative time updates

## Tasks / Subtasks

- [ ] Create Timestamp component (AC1, AC2)
  - [ ] Create `webapp/src/components/Timestamp.tsx`
  - [ ] Define TimestampProps interface
  - [ ] Implement component using React.memo for performance
  - [ ] Use useSelector to get current timezone from Redux store
  - [ ] Use useMemo for timestamp calculations
  - [ ] Convert Unix millis to moment object in user's timezone

- [ ] Implement format options (AC3)
  - [ ] Default format: moment 'lll' (Jan 18, 2026 10:30 AM)
  - [ ] Relative format: moment.fromNow() (5 minutes ago)
  - [ ] Custom format: Accept and use custom moment format strings
  - [ ] Add title attribute with full timestamp for hover tooltip

- [ ] Handle edge cases (AC4)
  - [ ] Check for 0 or null timestamp → display "Not yet decided"
  - [ ] Check for invalid timestamp (NaN, negative) → display "Invalid date"
  - [ ] Fallback timezone: use moment.tz.guess() if user timezone is null
  - [ ] Memoize calculations to prevent unnecessary re-renders

- [ ] Create unit tests (AC5)
  - [ ] Test component renders with valid Unix timestamp
  - [ ] Test timezone conversion (mock different timezones)
  - [ ] Test default format output
  - [ ] Test relative format output
  - [ ] Test custom format output
  - [ ] Test 0 timestamp displays "Not yet decided"
  - [ ] Test null timestamp displays "Not yet decided"
  - [ ] Test invalid timestamp displays "Invalid date"

- [ ] Remove HelloWorld component from Story 9.3
  - [ ] Delete `webapp/src/components/HelloWorld.tsx`
  - [ ] Remove HelloWorld registration from `webapp/src/index.tsx`

## Dev Notes

### Architecture Requirements

**Component Purpose:**
- Core reusable component for all approval timestamps
- Will be used in ApprovalPost, DM notifications, playbook posts
- Must be performant (memo + useMemo) as it appears multiple times per post

**Timezone Integration:**
- Redux selector: `getCurrentTimezone()` returns user's configured timezone
- Selector reactivity: Component auto-updates when timezone changes
- Fallback: Browser timezone via `moment.tz.guess()` if Redux returns null

### Component Implementation

```typescript
import React, {useMemo} from 'react';
import {useSelector} from 'react-redux';
import {getCurrentTimezone} from 'mattermost-redux/selectors/entities/timezone';
import moment from 'moment-timezone';

interface TimestampProps {
    unixMillis: number;
    format?: string;  // Default: 'lll'
    relative?: boolean;
}

const Timestamp: React.FC<TimestampProps> = React.memo(({
    unixMillis,
    format = 'lll',
    relative = false
}) => {
    const timezone = useSelector(getCurrentTimezone);

    const formattedTime = useMemo(() => {
        // Handle edge cases
        if (!unixMillis || unixMillis === 0) {
            return 'Not yet decided';
        }

        if (isNaN(unixMillis) || unixMillis < 0) {
            return 'Invalid date';
        }

        const tz = timezone || moment.tz.guess();
        const momentObj = moment.tz(unixMillis, tz);

        if (relative) {
            return momentObj.fromNow();
        }

        return momentObj.format(format);
    }, [unixMillis, timezone, format, relative]);

    const fullTimestamp = useMemo(() => {
        if (!unixMillis || unixMillis === 0 || isNaN(unixMillis)) {
            return '';
        }
        const tz = timezone || moment.tz.guess();
        return moment.tz(unixMillis, tz).format('LLLL z');
    }, [unixMillis, timezone]);

    return (
        <span title={fullTimestamp}>
            {formattedTime}
        </span>
    );
});

Timestamp.displayName = 'Timestamp';

export default Timestamp;
```

### Testing Requirements

**Unit Test Structure (using Jest + React Testing Library):**

```typescript
// webapp/src/components/Timestamp.test.tsx
import React from 'react';
import {render} from '@testing-library/react';
import {Provider} from 'react-redux';
import configureStore from 'redux-mock-store';
import Timestamp from './Timestamp';

const mockStore = configureStore([]);

describe('Timestamp Component', () => {
    const testTimestamp = 1705593000000; // Jan 18, 2024 10:30 AM UTC

    it('renders with default format', () => {
        const store = mockStore({
            entities: {
                timezone: {
                    automaticTimezone: 'America/Los_Angeles'
                }
            }
        });

        const {container} = render(
            <Provider store={store}>
                <Timestamp unixMillis={testTimestamp} />
            </Provider>
        );

        expect(container.textContent).toContain('Jan');
        expect(container.textContent).toContain('2024');
    });

    it('handles zero timestamp', () => {
        const store = mockStore({entities: {timezone: {}}});
        const {container} = render(
            <Provider store={store}>
                <Timestamp unixMillis={0} />
            </Provider>
        );

        expect(container.textContent).toBe('Not yet decided');
    });

    it('handles null timestamp', () => {
        const store = mockStore({entities: {timezone: {}}});
        const {container} = render(
            <Provider store={store}>
                <Timestamp unixMillis={null as any} />
            </Provider>
        );

        expect(container.textContent).toBe('Not yet decided');
    });

    it('handles invalid timestamp', () => {
        const store = mockStore({entities: {timezone: {}}});
        const {container} = render(
            <Provider store={store}>
                <Timestamp unixMillis={NaN} />
            </Provider>
        );

        expect(container.textContent).toBe('Invalid date');
    });

    it('displays relative time', () => {
        const store = mockStore({entities: {timezone: {}}});
        const fiveMinutesAgo = Date.now() - (5 * 60 * 1000);

        const {container} = render(
            <Provider store={store}>
                <Timestamp unixMillis={fiveMinutesAgo} relative />
            </Provider>
        );

        expect(container.textContent).toContain('minutes ago');
    });
});
```

### Library & Framework Requirements

**Dependencies (already installed):**
- react (React.memo, useMemo)
- react-redux (useSelector)
- moment-timezone (timezone conversion)
- mattermost-redux (getCurrentTimezone selector)

**New Dev Dependencies Needed:**
- @testing-library/react (unit testing)
- @testing-library/jest-dom (DOM matchers)
- redux-mock-store (mock Redux store for tests)
- jest (test runner)

**Add to webapp/package.json devDependencies:**
```json
"@testing-library/react": "^12.1.5",
"@testing-library/jest-dom": "^5.16.5",
"redux-mock-store": "^1.5.4",
"jest": "^29.5.0",
"@types/jest": "^29.5.0"
```

### File Structure Requirements

**Files to Create:**
- `webapp/src/components/Timestamp.tsx`
- `webapp/src/components/Timestamp.test.tsx`

**Files to Delete:**
- `webapp/src/components/HelloWorld.tsx` (verification component from 9.3)

**Files to Modify:**
- `webapp/src/index.tsx` (remove HelloWorld registration)
- `webapp/package.json` (add testing dependencies)

### References

- [Source: Epic 9 - Story 9.4 Acceptance Criteria]
- [Source: Epic 9 - Technology Stack] - moment-timezone, React 17
- [Moment.js Format Docs] - Format string options
- [Mattermost Redux Docs] - getCurrentTimezone selector

### Critical Gotchas

**AVOID THESE MISTAKES:**
1. **Don't skip React.memo**: Component used many times, memo prevents re-renders
2. **Don't skip useMemo**: Expensive timezone calculations, memoize results
3. **Don't assume timezone exists**: Always provide fallback
4. **Don't display raw timestamps**: Always format for readability

**Performance Considerations:**
- React.memo prevents re-renders when props unchanged
- useMemo caches formatted time until dependencies change
- Timezone selector updates component when user changes settings

## Dev Agent Record

### Agent Model Used

Claude Sonnet 4.5 (claude-sonnet-4-5-20250929)

### Debug Log References

N/A - Tests passed on first complete run

### Completion Notes List

**Implementation Decisions:**
1. Used `getUserTimezone(state, userId)` + `getCurrentUser()` instead of non-existent `getCurrentTimezone()` selector
2. Mattermost-redux v5.33 requires user ID for timezone lookup
3. Component checks `useAutomaticTimezone` flag and falls back to `manualTimezone` or browser timezone
4. Added `react-redux@^7.2.9` as production dependency (compatible with React 17)
5. Installed test dependencies: @testing-library/react, @testing-library/jest-dom, jest, ts-jest, redux-mock-store, jest-environment-jsdom
6. Created jest.config.js and jest.setup.js for test environment
7. Fixed NaN handling: Must check `isNaN()` before falsy checks (NaN || 0 = 0)
8. **Code Review Improvements:**
   - Extracted timezone resolution logic to `resolveTimezone()` helper (eliminated duplication)
   - Used proper `GlobalState` type instead of `any` for Redux state
   - Added timezone conversion accuracy tests (PST vs EST validation)
   - Added test for useAutomaticTimezone vs manualTimezone preference
   - Added test for timezone reactivity (Redux state changes)
   - Enhanced tooltip test to validate timezone abbreviation presence

**Test Results:**
- All 11 tests pass (100% coverage)
- Tests cover: default format, zero timestamp, null timestamp, NaN, negative, relative time, custom format, title attribute, timezone conversion accuracy, automatic vs manual timezone, Redux reactivity

**Performance:**
- React.memo applied to component (AC4)
- useMemo applied to formattedTime and fullTimestamp calculations (AC4)

### File List

**Files Created:**
- `webapp/src/components/Timestamp.tsx` (82 lines - with resolveTimezone helper)
- `webapp/src/components/Timestamp.test.tsx` (300 lines - 11 comprehensive tests)
- `webapp/jest.config.js` (20 lines - test configuration)
- `webapp/jest.setup.js` (1 line - test setup)

**Files Deleted:**
- `webapp/src/components/HelloWorld.tsx` (Story 9.3 verification component)

**Files Modified:**
- `webapp/src/index.tsx` (removed HelloWorld references, updated console message)
- `webapp/package.json` (added test dependencies: @testing-library/react, @testing-library/jest-dom, jest, ts-jest, redux-mock-store, jest-environment-jsdom, @types/jest, @types/redux-mock-store; added react-redux@^7.2.9; updated test script)
