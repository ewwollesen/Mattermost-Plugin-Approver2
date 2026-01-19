# Mattermost Approval Plugin - Webapp

React/TypeScript webapp component for the Mattermost Approval Workflow Plugin v3.0.0.

## Prerequisites

- Node.js v16+
- npm v8+

## Setup

Install dependencies:

```bash
npm install
```

## Development

Build webapp for development (with watch mode):

```bash
npm run dev
```

This runs webpack in watch mode and rebuilds on file changes.

## Production Build

Build optimized production bundle:

```bash
npm run build
```

Output: `dist/main.js` (minimized, ~584 bytes with externals)

## Code Quality

Run ESLint:

```bash
npm run lint
```

Format code with Prettier:

```bash
npm run format
```

## Architecture

### Technology Stack

- **React 17.0.2** - UI framework (Mattermost compatibility requirement)
- **TypeScript 4.9+** - Type safety with strict mode
- **Webpack 5** - Module bundler
- **moment-timezone** - Timezone conversion for approval timestamps

### Build Configuration

- **Entry Point:** `src/index.tsx` - Plugin initialization
- **Output:** `dist/main.js` - Webpack bundle
- **Externals:** React, ReactDOM, Redux, ReactRedux (provided by Mattermost - not bundled)
- **Source Maps:** Enabled for debugging

### Directory Structure

```
webapp/
├── src/
│   ├── components/     # React components (future stories)
│   ├── types/          # TypeScript type definitions (future stories)
│   └── index.tsx       # Plugin entry point with initialize() export
├── dist/               # Build output (gitignored)
├── package.json        # Dependencies and scripts
├── tsconfig.json       # TypeScript configuration
├── webpack.config.js   # Webpack build configuration
├── .eslintrc.js        # ESLint rules
└── .prettierrc         # Prettier formatting rules
```

## Plugin Integration

The webapp is loaded by Mattermost when the plugin starts. The entry point (`src/index.tsx`) exports an `initialize(registry, store)` function that Mattermost calls to register components.

### Registration (Story 9.2+)

```typescript
export function initialize(registry: any, store: any) {
    // Component registration will be added in Story 9.2+
    registry.registerPostTypeComponent('custom_approval', ApprovalPost);
}
```

## Version

**Current:** v3.0.0 (Epic 9: Webapp Component Framework)

## Notes

- No unit tests in Story 9.1 (infrastructure setup only)
- Component testing starts in Story 9.4 (Timestamp component)
- Build integrated with project Makefile (see root README)
