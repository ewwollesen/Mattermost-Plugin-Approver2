# Story 7.2: Fix CI/CD Build Failures on Tag Push

**Epic:** 7 - 1.0 Polish & UX Improvements
**Story ID:** 7.2
**Priority:** CRITICAL (Release Blocker)
**Status:** ready-for-dev
**Created:** 2026-01-15

---

## User Story

**As a** maintainer
**I want** tag pushes to trigger clean builds without failures
**So that** I can release with confidence

---

## Story Context

This is a CRITICAL release blocker for 1.0. Currently, when pushing a tag to master (e.g., `v1.0.0`), the CI/CD pipeline triggers multiple builds, two of which consistently fail. This prevents reliable releases and undermines confidence in the deployment process.

**Business Impact:**
- Cannot ship 1.0 release until this is resolved
- Failed builds create confusion about release readiness
- Wastes CI/CD minutes and maintainer time investigating failures
- Blocks the entire Epic 7 completion

**Technical Context:**
After manual investigation, the root cause has been identified: **BOTH `.github/workflows/ci.yml` AND `.github/workflows/release.yml` trigger on tag push**, creating duplicate/conflicting build processes.

---

## Acceptance Criteria

### AC1: Single Successful Build on Tag Push
**Given** I push a version tag (e.g., `v1.0.0`) to the master branch
**When** the CI/CD pipeline executes
**Then** ONLY the necessary workflows trigger
**And** ALL triggered workflows complete successfully with zero failures
**And** build logs are clean and actionable

### AC2: Redundant Build Triggers Eliminated
**Given** the CI/CD configuration files (.github/workflows/)
**When** a tag is pushed to master
**Then** no duplicate or redundant workflows execute
**And** the ci.yml workflow does NOT trigger on tag push (only on branch push and PRs)
**And** the release.yml workflow ONLY triggers on tag push

### AC3: Release Workflow Documented
**Given** the fix has been implemented
**When** a maintainer needs to create a release
**Then** the release process is documented clearly
**And** the documentation explains what triggers each workflow
**And** the maintainer knows exactly which workflow handles releases

### AC4: Verified with Test Tag
**Given** the CI/CD fix is complete
**When** I push a test tag (e.g., `v0.9.9-test`) to verify the fix
**Then** the release workflow triggers successfully
**And** release artifacts are generated correctly
**And** no unexpected workflows trigger or fail

---

## Current Problem Analysis

### Current CI/CD Configuration

**File: `.github/workflows/ci.yml`**
```yaml
name: ci
on:
  schedule:
    - cron: "0 0 * * *"
  push:
    branches:
      - master
    tags:
      - "v*"  # ❌ PROBLEM: This triggers ci workflow on tag push
  pull_request:
```

**File: `.github/workflows/release.yml`**
```yaml
name: Release
on:
  push:
    tags:
      - 'v*'  # ✅ This is correct - release workflow should trigger on tags
```

### Root Cause Identified

**When a tag like `v1.0.0` is pushed:**

1. **ci.yml triggers** (line 8-9: `tags: - "v*"`)
   - Calls `mattermost/actions-workflows/.github/workflows/plugin-ci.yml@main`
   - This is a reusable workflow that runs tests, linting, and builds
   - May fail because it's designed for PR/branch validation, not release builds

2. **release.yml triggers** (line 5-6: `tags: - 'v*'`)
   - Runs `make dist` to build plugin tarball
   - Creates GitHub Release with artifacts
   - This is the CORRECT workflow for tag push

3. **Result:** Multiple workflows run simultaneously, causing:
   - Resource contention (both trying to build at same time)
   - Environment conflicts (different Go setups, different contexts)
   - Redundant work (building the plugin twice)
   - **TWO FAILURES** from the duplicated/conflicting builds

### Why This Happens

The `ci.yml` workflow was designed for:
- **Scheduled runs** (daily cron)
- **Branch pushes** (continuous integration on master)
- **Pull requests** (PR validation)

It was INCORRECTLY configured to also run on **tag pushes**, which is the domain of the release workflow.

---

## Technical Implementation Plan

### Fix Strategy: Remove Tag Trigger from ci.yml

**Simple, surgical fix:**
1. Edit `.github/workflows/ci.yml`
2. Remove the `tags` section from the `on.push` trigger
3. Keep ci.yml for branch/PR validation only
4. Let release.yml handle ALL tag-related builds

### Updated ci.yml Configuration

**Before:**
```yaml
on:
  schedule:
    - cron: "0 0 * * *"
  push:
    branches:
      - master
    tags:
      - "v*"  # ❌ Remove this
  pull_request:
```

**After:**
```yaml
on:
  schedule:
    - cron: "0 0 * * *"
  push:
    branches:
      - master
    # tags section REMOVED - releases handled by release.yml
  pull_request:
```

### Verification Steps

After making the fix:

1. **Commit the change:**
   ```bash
   git add .github/workflows/ci.yml
   git commit -m "Fix CI/CD: Remove tag trigger from ci.yml to prevent duplicate builds

   - ci.yml now only triggers on branch push and PR
   - release.yml exclusively handles tag-based releases
   - Fixes duplicate build issue causing 2 failures on tag push

   Story 7.2: Fix CI/CD Build Failures on Tag Push

   Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
   ```

2. **Push to master:**
   ```bash
   git push origin master
   ```

3. **Create and push a test tag:**
   ```bash
   git tag v0.9.9-test
   git push origin v0.9.9-test
   ```

4. **Verify in GitHub Actions:**
   - Go to: https://github.com/ewwollesen/Mattermost-Plugin-Approver2/actions
   - Confirm ONLY the "Release" workflow triggered
   - Confirm "ci" workflow did NOT trigger
   - Confirm the "Release" workflow completed successfully
   - Confirm release artifacts were created

5. **Clean up test tag:**
   ```bash
   git tag -d v0.9.9-test
   git push origin :refs/tags/v0.9.9-test
   gh release delete v0.9.9-test --yes
   ```

---

## Implementation Requirements

### Files to Modify

**`.github/workflows/ci.yml`**
- Location: `.github/workflows/ci.yml`
- Action: Remove lines 8-9 (the `tags` section)
- Preserve: All other triggers (schedule, branch push, pull_request)

### Testing Requirements

1. **Pre-push validation:**
   - Verify yaml syntax is valid: `yamllint .github/workflows/ci.yml`
   - Ensure no indentation errors
   - Confirm the on.push.branches section remains intact

2. **Live validation:**
   - Push test tag and monitor GitHub Actions UI
   - Verify only release.yml triggers
   - Verify release.yml completes successfully
   - Verify artifacts are generated correctly

3. **Regression check:**
   - After fix, push a commit to master (not a tag)
   - Verify ci.yml DOES trigger on branch push
   - Confirm PR validation still works

### Documentation Updates

**Update `.github/workflows/ci.yml` with comment:**
```yaml
name: ci
# This workflow runs on:
# - Daily schedule (cron)
# - Pushes to master branch
# - Pull requests
# Note: Tag-based releases are handled exclusively by release.yml
on:
  schedule:
    - cron: "0 0 * * *"
  push:
    branches:
      - master
    # Tags are NOT included here - see release.yml for tag-based builds
  pull_request:
```

---

## Architecture Compliance

### NFR-M1: Code Simplicity
✅ This fix follows the principle of simplicity - remove redundant configuration rather than adding complexity.

### NFR-M4: Clear Error Messages
✅ The commit message clearly explains what changed and why, making the fix auditable.

### Mattermost Plugin Standards
✅ GitHub Actions workflows are standard for Mattermost plugins (see mattermost-plugin-starter-template).

---

## Developer Notes

### Why Separate CI and Release Workflows?

**ci.yml Purpose:**
- Continuous Integration: validate code quality on every commit
- Run tests, linting, type checking
- Fast feedback for developers
- Runs on: branches, PRs, scheduled

**release.yml Purpose:**
- Release Management: build and publish versioned artifacts
- Create GitHub releases with changelogs
- Build production-ready tarballs
- Runs on: version tags only

**Separation Benefits:**
- Clear responsibilities
- No resource contention
- Different environments/permissions
- Easier to debug issues

### Common Pitfall to Avoid

**DON'T do this:**
```yaml
# BAD: Trying to use ci.yml for both CI and releases
on:
  push:
    branches:
      - master
    tags:
      - "v*"  # ❌ Creates confusion and conflicts
```

**DO this:**
```yaml
# GOOD: Separate concerns clearly
# ci.yml: branches and PRs only
on:
  push:
    branches:
      - master
  pull_request:

# release.yml: tags only
on:
  push:
    tags:
      - 'v*'
```

### Alternative Solutions Considered

**Option A: Conditional Jobs (REJECTED)**
```yaml
jobs:
  ci:
    if: ${{ !startsWith(github.ref, 'refs/tags/') }}
```
- More complex
- Harder to debug
- Still triggers workflow unnecessarily

**Option B: Remove ci.yml Tag Trigger (SELECTED)**
- Simple, surgical fix
- Clear separation of concerns
- Follows Mattermost plugin patterns
- Easy to understand and maintain

---

## Related Context

### Recent Commits Pattern
```
2a38ea6 Epic 6: Request Timeout & Verification (Feature Complete for 1.0)
ae5fd29 Story 5.2: Code review fixes - Update help documentation
cde3ee3 Epic 5: List Filtering and Default View Optimization
```

Pattern: Each epic completion likely triggered a tag release, exposing this CI/CD issue.

### Project Build System
- **Build tool:** Makefile (see root Makefile)
- **Go version:** 1.24 (specified in both ci.yml and release.yml)
- **Plugin packaging:** `make dist` creates tarball in `dist/` directory
- **Artifact format:** `{PLUGIN_ID}-{VERSION}.tar.gz`

### Testing Infrastructure
- **Test command:** `make test`
- **Test flags:** GO_TEST_FLAGS ?= -race (see Makefile line 6)
- **Testing approach:** Standard Go testing with testify, table-driven tests
- **Coverage expectation:** Minimum 70% for core logic (NFR-M2)

---

## Definition of Done

- [x] `.github/workflows/ci.yml` updated to remove tag trigger
- [x] Explanatory comment added to ci.yml documenting the change
- [x] Changes committed with clear commit message
- [x] Changes pushed to master branch
- [x] Test tag created and pushed (e.g., `v0.9.9-test`)
- [x] Verified: Only release.yml triggered on test tag push
- [x] Verified: ci.yml did NOT trigger on test tag push
- [x] Verified: Release workflow completed successfully (zero failures)
- [x] Verified: Release artifacts generated correctly
- [x] Test tag and release cleaned up
- [x] Regression test: ci.yml still triggers on branch push to master
- [x] Regression test: ci.yml still triggers on pull request
- [x] No build failures exist
- [x] This story marked as done in sprint-status.yaml

---

## Success Criteria

**Primary Metric:** Zero CI/CD failures on tag push
**Validation:** Push `v1.0.0` tag and see 1 successful build (release.yml only)

**Before this fix:**
- Push tag → 3 workflows trigger → 2 fail → ❌ Cannot release

**After this fix:**
- Push tag → 1 workflow triggers (release.yml) → 0 failures → ✅ Clean release

---

## Notes

- This is the FIRST story to implement in Epic 7 (per implementation order)
- Must be completed before 7.1 (cancellation notification) can be tested in CI
- Estimated time: 30 minutes (simple config change + verification)
- Risk: Very low (surgical fix, easily reversible)
- Testing: Can be validated immediately with test tag

---

## Dev Agent Record

### Implementation Summary
- Removed tag trigger from `.github/workflows/ci.yml` (lines 8-9)
- Added explanatory comments documenting workflow separation
- Fixed 5 linting issues discovered during CI validation:
  - 3 gofmt formatting issues (test files)
  - 1 govet variable shadowing (server/plugin.go:366)
  - 1 staticcheck empty branch (server/approval/service.go:170)
  - 1 modernize linter (interface{} → any in checker_test.go)
- Cleaned up go.mod/go.sum (removed unused mattermost-plugin-starter-template dependency)

### Commits
1. `87b05c1` - Fix CI/CD: Remove tag trigger from ci.yml to prevent duplicate builds
2. `b936a5d` - Fix linting errors: gofmt, govet, staticcheck
3. `cab3497` - Fix final linting issue: modernize interface{} to any
4. `b69793a` - Clean up go.mod: remove unused starter-template dependency

### Verification Results
- Test tag `v0.9.9-test` pushed successfully
- Verified: Only release.yml triggered on tag push ✓
- Verified: ci.yml did NOT trigger on tag push ✓
- Verified: Release workflow completed successfully ✓
- Verified: ci.yml still triggers on master branch push ✓
- All CI checks passing (green build) ✓
- Test tag and release cleaned up ✓

### Technical Notes
Root cause: Both ci.yml and release.yml configured to trigger on tag push, causing duplicate builds and failures. Simple fix: removed tags section from ci.yml, letting release.yml exclusively handle tag-based releases.

Additional work: Fixed cascading linting issues discovered during CI validation. Final blocker was modernize linter requiring `interface{}` → `any` conversion.

---

## File List

### Modified Files
- `.github/workflows/ci.yml` - Removed tag trigger, added documentation comments
- `server/approval/service.go` - Fixed staticcheck empty branch warning
- `server/approval/service_test.go` - Fixed gofmt formatting
- `server/notifications/dm_test.go` - Fixed gofmt formatting
- `server/plugin.go` - Fixed govet variable shadowing
- `server/timeout/checker_test.go` - Fixed gofmt formatting, modernized interface{} to any
- `go.mod` - Removed unused starter-template dependency
- `go.sum` - Updated checksums

---

## Change Log

**2026-01-15:** Story 7.2 completed
- Fixed CI/CD duplicate build issue by removing tag trigger from ci.yml
- Resolved all linting errors blocking CI pipeline
- Cleaned up unused dependencies
- All acceptance criteria satisfied, CI green

---

**Story Status:** done
**Completed:** 2026-01-15
