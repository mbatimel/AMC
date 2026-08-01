# Admin-Front Split Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Extract the admin panel currently living inside `front` (`/admin/*`) into a standalone Next.js project `admin-front`, buildable and deployable on its own domain (`admin.wk.amctechgroup.ru`), while `front` keeps serving the public site and stays the single source of truth for the temporary `portal-api` data layer.

**Architecture:** Two independent Next.js 16 apps side by side in the repo (`front/`, `admin-front/`). Admin-only code (`adminSession` entity, `admin.ts`/`portalUsers.ts`/`signupRequests.ts` API clients, admin views/routes) moves out of `front` entirely. Code used by both apps (`content`/`feedback`/`support`/`products` API clients, `AuthShell`/`FormSelect`/`Toast` UI, a few icons/lib helpers, styles) is duplicated by copy — no shared package, no workspaces. `portal-api` route handlers and the file-backed store stay only in `front`; `admin-front` reaches them through nginx (prod) and a `next.config.ts` rewrite fallback (local dev). Admin routes drop the `/admin` prefix since they now live on their own domain.

**Tech Stack:** Next.js 16, React 19, TypeScript 6, effector/effector-react, `@heroui/react`, Tailwind CSS v4, ESLint 9/typescript-eslint 8, Prettier 3, Stylelint 17, Yarn, Docker (standalone output), nginx.

## Global Constraints

- No test framework exists in `front` (no `test` script, no Jest/Vitest). Verification for every task is `yarn typecheck`, `yarn lint`, and `yarn build` — run these instead of a test suite.
- Package manager is Yarn (classic, `yarn.lock` present) — use `yarn`, not `npm`/`pnpm`.
- `admin-front` must NOT contain `src/app/portal-api/*` or the stateful parts of `src/core/shared/server/portal/*` (`store.ts`, `defaults.ts`, `response.ts`) — that stays exclusively in `front` (design decision, avoids split-brain state in `portal.json`). **Correction found during Task 2 review:** `src/core/shared/server/portal/types.ts` is the one exception — it's a pure type-definitions file (158 lines, zero imports, no Node/fs code), and `admin.ts`, `portalUsers.ts`, `signupRequests.ts`, `content.ts`, `feedback.ts`, and `support.ts` all import types from it. It must be duplicated into `admin-front/src/core/shared/server/portal/types.ts` (same copy-by-value pattern as the API clients) — see the Task 2 correction below.
- **Correction found during Task 2 review:** `signupRequests.ts` is not admin-only — `front/src/core/entities/session/model.ts` (`registerModerationFx`, the public registration flow) calls `createSignupRequest` from it after a customer signs up, so the public site needs it too. It belongs in the **duplicate** bucket (like `content.ts`/`feedback.ts`/`support.ts`/`products.ts`), not the **move** bucket. Task 2 below still moves it into `admin-front` (needed there for `listSignupRequests`/`decideSignupRequest`), but a fix round restores a copy at `front/src/core/shared/api/signupRequests.ts` too.
- Admin routes in `admin-front` are de-prefixed: `/admin/products` → `/products`, `/admin/login` → `/login`, admin dashboard `/admin` → `/`.
- Cookie name stays `admin_user_id` (`ADMIN_USER_ID_COOKIE`), host-only (no `Domain=` attribute) — no change needed, it will naturally scope to `admin.wk.amctechgroup.ru`.
- Every file copy must preserve file content exactly except where a task explicitly lists a rename/edit.

---

## Task 1: Scaffold the admin-front project

**Files:**
- Create: `admin-front/package.json`
- Create: `admin-front/tsconfig.json`
- Create: `admin-front/next-env.d.ts`
- Create: `admin-front/next.config.ts`
- Create: `admin-front/eslint.config.js`
- Create: `admin-front/prettier.config.js`
- Create: `admin-front/postcss.config.js`
- Create: `admin-front/stylelint.config.js`
- Create: `admin-front/.gitignore`
- Create: `admin-front/.prettierignore`
- Create: `admin-front/.cursorignore`
- Create: `admin-front/.env.example`
- Create: `admin-front/public/.gitkeep`

**Interfaces:**
- Produces: a working, empty Next.js app skeleton at `admin-front/` that later tasks add `src/` to. `yarn` in `admin-front/` must succeed before Task 2 starts.

- [ ] **Step 1: Create directories**

```bash
mkdir -p admin-front/src admin-front/public
```

- [ ] **Step 2: Write `admin-front/package.json`**

```json
{
  "name": "amc-admin-front",
  "private": true,
  "type": "module",
  "scripts": {
    "dev": "next dev",
    "build": "next build",
    "start": "next start",
    "lint": "yarn lint:js && yarn lint:css",
    "lint:fix": "yarn lint:js:fix && yarn lint:css:fix",
    "lint:js": "eslint \"src/**/*.{ts,tsx}\" \"next.config.ts\" --no-error-on-unmatched-pattern",
    "lint:js:fix": "eslint \"src/**/*.{ts,tsx}\" \"next.config.ts\" --fix --no-error-on-unmatched-pattern",
    "lint:css": "stylelint \"src/**/*.css\"",
    "lint:css:fix": "stylelint \"src/**/*.css\" --fix",
    "format": "prettier --write .",
    "format:check": "prettier --check .",
    "prepare": "cd .. && husky .husky",
    "typecheck": "tsc --noEmit"
  },
  "lint-staged": {
    "src/**/*.{ts,tsx}": [
      "eslint --fix --no-error-on-unmatched-pattern",
      "prettier --write"
    ],
    "src/**/*.css": [
      "stylelint --fix",
      "prettier --write"
    ],
    "next.config.ts": [
      "eslint --fix --no-error-on-unmatched-pattern",
      "prettier --write"
    ],
    "**/*.{js,json,md}": "prettier --write"
  },
  "devDependencies": {
    "@eslint/js": "^10.0.1",
    "@tailwindcss/postcss": "^4.3.1",
    "@types/node": "^26.0.1",
    "@types/react": "^19.2.17",
    "@types/react-dom": "^19.2.3",
    "eslint": "^9.39.0",
    "eslint-config-prettier": "^10.1.8",
    "eslint-plugin-effector": "^0.19.0",
    "eslint-plugin-perfectionist": "^5.9.1",
    "eslint-plugin-react": "^7.37.5",
    "eslint-plugin-react-hooks": "^7.1.1",
    "globals": "^17.7.0",
    "husky": "^9.1.7",
    "lint-staged": "^17.2.0",
    "postcss": "^8.5.15",
    "postcss-mixins": "^12.1.2",
    "prettier": "^3.9.1",
    "react-scan": "0.5.7",
    "stylelint": "^17.14.0",
    "stylelint-config-standard": "^40.0.0",
    "stylelint-use-logical": "^2.1.3",
    "tailwindcss": "^4.3.1",
    "typescript": "^6.0.3",
    "typescript-eslint": "^8.62.0"
  },
  "dependencies": {
    "@heroui/react": "^3.2.1",
    "@heroui/styles": "^3.2.1",
    "clsx": "^2.1.1",
    "effector": "^23.4.4",
    "effector-react": "^23.3.0",
    "next": "^16.2.9",
    "react": "^19.2.7",
    "react-dom": "^19.2.7"
  },
  "resolutions": {
    "react-grab": "0.1.50"
  }
}
```

Version numbers copied verbatim from `front/package.json`. `libphonenumber-js` is intentionally dropped — nothing admin-side uses it (only `core/shared/lib/validateContact.ts`, which isn't part of the admin duplication set).

- [ ] **Step 3: Write `admin-front/tsconfig.json`** (identical to `front/tsconfig.json`)

```json
{
  "compilerOptions": {
    "target": "ES2022",
    "lib": ["dom", "dom.iterable", "esnext"],
    "allowJs": true,
    "skipLibCheck": true,
    "strict": true,
    "noEmit": true,
    "esModuleInterop": true,
    "module": "esnext",
    "moduleResolution": "bundler",
    "resolveJsonModule": true,
    "isolatedModules": true,
    "jsx": "react-jsx",
    "incremental": true,
    "noUnusedLocals": true,
    "noUnusedParameters": true,
    "noFallthroughCasesInSwitch": true,
    "plugins": [
      {
        "name": "next"
      }
    ],
    "paths": {
      "@/*": ["./src/*"]
    }
  },
  "include": [
    "next-env.d.ts",
    "src/**/*.ts",
    "src/**/*.tsx",
    "next.config.ts",
    ".next/types/**/*.ts",
    ".next/dev/types/**/*.ts"
  ],
  "exclude": ["node_modules"]
}
```

(`vi-portal` excluded in front's tsconfig doesn't exist here, so it's dropped from `exclude`.)

- [ ] **Step 4: Write `admin-front/next-env.d.ts`**

```ts
/// <reference types="next" />
/// <reference types="next/image-types/global" />
import "./.next/dev/types/routes.d.ts";

// NOTE: This file should not be edited
// see https://nextjs.org/docs/app/api-reference/config/typescript for more information.
```

- [ ] **Step 5: Write `admin-front/next.config.ts`**

```ts
import type { NextConfig } from 'next';

const apiProxyTarget = process.env.API_PROXY_TARGET ?? 'https://wk.amctechgroup.ru';
const portalApiProxyTarget =
  process.env.PORTAL_API_PROXY_TARGET ?? 'https://wk.amctechgroup.ru';

const nextConfig: NextConfig = {
  async rewrites() {
    return [
      {
        destination: `${apiProxyTarget}/api/v1/orders/`,
        source: '/api/v1/orders',
      },
      {
        destination: `${apiProxyTarget}/api/:path*`,
        source: '/api/:path*',
      },
      {
        destination: `${portalApiProxyTarget}/portal-api/:path*`,
        source: '/portal-api/:path*',
      },
    ];
  },
  skipTrailingSlashRedirect: true,
  output: 'standalone',
  reactStrictMode: true,
};

export default nextConfig;
```

This mirrors `front/next.config.ts` and adds the `/portal-api/:path*` fallback rewrite to `front`'s domain, needed for `yarn dev` without nginx in front of it. In production nginx handles `/portal-api/` directly (Task 9), so this rewrite is a dev-only safety net there too — but harmless to keep in both places.

- [ ] **Step 6: Write remaining config files, verbatim copies of `front`'s**

`admin-front/eslint.config.js` — identical content to `front/eslint.config.js`, except drop `'vi-portal/**'` from `globalIgnores` (that path doesn't exist in this project):

```js
import js from '@eslint/js';
import { defineConfig, globalIgnores } from 'eslint/config';
import eslintConfigPrettier from 'eslint-config-prettier';
import effector from 'eslint-plugin-effector';
import perfectionist from 'eslint-plugin-perfectionist';
import react from 'eslint-plugin-react';
import reactHooks from 'eslint-plugin-react-hooks';
import globals from 'globals';
import tseslint from 'typescript-eslint';

export default defineConfig([
  globalIgnores([
    '.next/**',
    'build/**',
    'dist/**',
    'eslint.config.js',
    'next.config.ts',
    'node_modules/**',
    'postcss.config.js',
    'prettier.config.js',
    'stylelint.config.js',
  ]),
  js.configs.recommended,
  ...tseslint.configs.recommendedTypeChecked.map((config) => ({
    ...config,
    files: ['**/*.{ts,tsx}'],
  })),
  {
    files: ['**/*.{ts,tsx}'],
    languageOptions: {
      globals: globals.browser,
      parserOptions: {
        projectService: true,
        tsconfigRootDir: import.meta.dirname,
      },
    },
    plugins: {
      react,
      'react-hooks': reactHooks,
    },
    rules: {
      ...react.configs.flat.recommended.rules,
      ...react.configs.flat['jsx-runtime'].rules,
      'import/order': 'off',
      'react/react-in-jsx-scope': 'off',
    },
    settings: {
      react: {
        version: 'detect',
      },
    },
  },
  effector.flatConfigs.recommended,
  effector.flatConfigs.react,
  perfectionist.configs['recommended-natural'],
  eslintConfigPrettier,
]);
```

`admin-front/prettier.config.js` (identical to `front`'s):

```js
/** @type {import('prettier').Config} */
export default {
  printWidth: 100,
  singleQuote: true,
  trailingComma: 'all',
};
```

`admin-front/postcss.config.js` (identical to `front`'s):

```js
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import tailwindcss from '@tailwindcss/postcss';
import postcssMixins from 'postcss-mixins';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

/** @type {import('postcss-load-config').Config} */
export default {
  plugins: [
    tailwindcss,
    postcssMixins({
      mixinsDir: path.join(__dirname, 'src/core/shared/styles/mixins'),
    }),
  ],
};
```

`admin-front/stylelint.config.js` (identical to `front`'s, minus the `vi-portal` ignore):

```js
/** @type {import('stylelint').Config} */
export default {
  extends: ['stylelint-config-standard'],
  plugins: ['stylelint-use-logical'],
  ignoreFiles: ['**/node_modules/**', '**/.next/**', '**/dist/**', '**/build/**', '**/out/**'],
  rules: {
    'at-rule-no-unknown': [
      true,
      {
        ignoreAtRules: ['define-mixin', 'mixin', 'mixin-content'],
      },
    ],
    'csstools/use-logical': 'always',
    'import-notation': 'string',
    'media-feature-range-notation': 'context',
    'property-no-unknown': [
      true,
      {
        ignoreProperties: ['composes'],
      },
    ],
    'selector-pseudo-class-no-unknown': [
      true,
      {
        ignorePseudoClasses: ['global'],
      },
    ],
  },
  overrides: [
    {
      files: ['src/**/*.module.css'],
      rules: {
        'selector-class-pattern': [
          '^[a-z][a-zA-Z0-9]*$',
          {
            message: 'Expected class selector to be camelCase',
          },
        ],
      },
    },
    {
      files: ['src/core/shared/styles/mixins/**/*.css'],
      rules: {
        'nesting-selector-no-missing-scoping-root': null,
      },
    },
    {
      files: ['src/core/shared/styles/normalize.css'],
      rules: {
        'csstools/use-logical': null,
      },
    },
  ],
};
```

- [ ] **Step 7: Write ignore/env files**

`admin-front/.gitignore`:

```
dist
.next
out
node_modules/
*.local
tsconfig.tsbuildinfo
```

`admin-front/.prettierignore`:

```
node_modules/
dist/
build/
yarn.lock
```

`admin-front/.cursorignore`:

```
node_modules/
dist/
```

`admin-front/.env.example`:

```
# Базовый адрес backend-сервисов (products/users/orders/access/admin)
API_PROXY_TARGET=https://wk.amctechgroup.ru

# Адрес, откуда проксируется /portal-api (временный слой контента/support/
# feedback/portal-users/signup-requests/audit-log — источник правды в front)
PORTAL_API_PROXY_TARGET=https://wk.amctechgroup.ru
```

- [ ] **Step 8: Create empty public dir placeholder**

```bash
touch admin-front/public/.gitkeep
```

- [ ] **Step 9: Verify install**

Run: `cd admin-front && yarn install`
Expected: completes without error, creates `admin-front/node_modules` and `admin-front/yarn.lock`.

- [ ] **Step 10: Commit**

```bash
git add admin-front
git commit -m "chore(admin-front): scaffold standalone Next.js project"
```

---

## Task 2: Move admin-only entity and API clients

Moves code used **only** by the admin panel — nothing in `front` outside `views/Admin`/`app/admin` references these, so they're deleted from `front`, not duplicated.

**Files:**
- Move: `front/src/core/entities/adminSession/` → `admin-front/src/core/entities/adminSession/`
- Move: `front/src/core/shared/api/admin.ts` → `admin-front/src/core/shared/api/admin.ts`
- Move: `front/src/core/shared/api/portalUsers.ts` → `admin-front/src/core/shared/api/portalUsers.ts`
- Move: `front/src/core/shared/api/signupRequests.ts` → `admin-front/src/core/shared/api/signupRequests.ts`

**Interfaces:**
- Consumes: nothing yet (Task 3 provides `portalClient.ts`/`parseApiError.ts` that `admin.ts` imports from — do Task 3 before running `yarn build` on `admin-front`, but the `git mv` in this task can happen first).
- Produces: `@/core/entities/adminSession` (`useAdminSession`, `ADMIN_USER_ID_COOKIE`, `ADMIN_ACTOR_LABEL`, `$adminUserId` etc.) and `@/core/shared/api/admin` (`adminLoginRequest`, `adminLogoutRequest`, `adminSessionRequest`, `adminAuditLogRequest`, `AdminApiError`, `AdminSessionResponse` type) for later tasks.

- [ ] **Step 1: Move the files with git mv**

```bash
mkdir -p admin-front/src/core/entities admin-front/src/core/shared/api
git mv front/src/core/entities/adminSession admin-front/src/core/entities/adminSession
git mv front/src/core/shared/api/admin.ts admin-front/src/core/shared/api/admin.ts
git mv front/src/core/shared/api/portalUsers.ts admin-front/src/core/shared/api/portalUsers.ts
git mv front/src/core/shared/api/signupRequests.ts admin-front/src/core/shared/api/signupRequests.ts
```

- [ ] **Step 2: Verify no leftover references in front**

Run: `grep -rn "adminSession\|core/shared/api/admin'\|core/shared/api/portalUsers'\|core/shared/api/signupRequests'" front/src`
Expected: no output (Task 7 removes the last consumers — `views/Admin`/`app/admin` — so this check is expected to still show hits from those until Task 7; that's fine, just confirm the four moved paths themselves no longer exist under `front/src`).

Run: `ls front/src/core/entities/adminSession front/src/core/shared/api/admin.ts 2>&1`
Expected: `No such file or directory` for both.

- [ ] **Step 3: Commit**

```bash
git add -A
git commit -m "refactor: move adminSession entity and admin-only API clients to admin-front"
```

---

## Task 3: Duplicate shared API-layer files into admin-front

These files are used by **both** apps (public site reads/writes the same `portal-api` data that admin edits), so they're copied, not moved — `front` keeps its own copy unchanged.

**Correction found during Task 2 review:** `content.ts`, `feedback.ts`, and `support.ts` each import types from `@/core/shared/server/portal/types` (`ContentPageKey`/etc., `OrderFeedback`, `SupportRequest`/`SupportRequestStatus`). That file must exist in `admin-front` for these to compile. Task 2's fix round already copied `admin-front/src/core/shared/server/portal/types.ts` (needed there for `admin.ts`/`portalUsers.ts`/`signupRequests.ts` too) — by the time this task runs, that file already exists in `admin-front`. Don't copy it again; just confirm it's present (Step 0 below).

**Files:**
- Create (copy of `front/src/core/shared/api/content.ts`): `admin-front/src/core/shared/api/content.ts`
- Create (copy of `front/src/core/shared/api/feedback.ts`): `admin-front/src/core/shared/api/feedback.ts`
- Create (copy of `front/src/core/shared/api/support.ts`): `admin-front/src/core/shared/api/support.ts`
- Create (copy of `front/src/core/shared/api/products.ts`): `admin-front/src/core/shared/api/products.ts`
- Create (copy of `front/src/core/shared/api/portalClient.ts`): `admin-front/src/core/shared/api/portalClient.ts`
- Create (copy of `front/src/core/shared/api/parseApiError.ts`): `admin-front/src/core/shared/api/parseApiError.ts`

**Interfaces:**
- Produces: `@/core/shared/api/{content,feedback,support,products,portalClient,parseApiError}` in admin-front, byte-identical to front's versions at copy time. `admin.ts`, `portalUsers.ts`, `signupRequests.ts` (moved in Task 2) import `portalRequest`/`PortalApiError` from `./portalClient` and error helpers from `./parseApiError` — this task unblocks their compilation.

- [ ] **Step 0: Confirm `server/portal/types.ts` is already present in admin-front**

Run: `ls admin-front/src/core/shared/server/portal/types.ts`
Expected: file exists (copied there by Task 2's fix round). If it's missing, copy it now: `mkdir -p admin-front/src/core/shared/server/portal && cp front/src/core/shared/server/portal/types.ts admin-front/src/core/shared/server/portal/types.ts` — and nothing else from `front/src/core/shared/server/` (no `store.ts`, `defaults.ts`, `response.ts`, no `server/assistant/*`).

- [ ] **Step 1: Copy the files**

```bash
cp front/src/core/shared/api/content.ts admin-front/src/core/shared/api/content.ts
cp front/src/core/shared/api/feedback.ts admin-front/src/core/shared/api/feedback.ts
cp front/src/core/shared/api/support.ts admin-front/src/core/shared/api/support.ts
cp front/src/core/shared/api/products.ts admin-front/src/core/shared/api/products.ts
cp front/src/core/shared/api/portalClient.ts admin-front/src/core/shared/api/portalClient.ts
cp front/src/core/shared/api/parseApiError.ts admin-front/src/core/shared/api/parseApiError.ts
```

- [ ] **Step 2: Verify byte-for-byte copies**

Run: `diff front/src/core/shared/api/content.ts admin-front/src/core/shared/api/content.ts && diff front/src/core/shared/api/products.ts admin-front/src/core/shared/api/products.ts`
Expected: no output (files identical).

- [ ] **Step 3: Commit**

```bash
git add admin-front/src/core/shared/api
git commit -m "chore(admin-front): duplicate shared API clients from front"
```

---

## Task 4: Duplicate shared UI kit, icons, lib helpers, and styles

**Correction found while preparing this task's dispatch (same class of gap as Tasks 2-3 — a dependency the original file list missed):** `AuthShell` is not a drop-in copy like the API clients were. Its `front` version hard-imports three things that don't exist in `admin-front`:
1. `authErrorCleared` from `@/core/entities/session` (the **public** session entity — `admin-front` has no such entity, only `adminSession`). It's called on mount to clear the public login form's error state; the admin-front copy must clear the admin session's error state instead, using `adminAuthErrorCleared` — already exported from `@/core/entities/adminSession` (moved there in Task 2).
2. `logoImage`/`logoImage2x` from `@/core/shared/ui/Header/assets/logo.png` — `admin-front` has no `Header` component (never copied, not needed). The two PNG files need to move to a location admin-front actually has.
3. `FormSelect` imports `IconChevronDown` from the `@/core/shared/icons` **barrel** (`import { IconChevronDown } from '@/core/shared/icons'`) — `admin-front` doesn't have that icon file or a barrel at all yet (only `IconKey.tsx` was planned).

This task now includes copying `IconChevronDown.tsx`, creating a minimal icons barrel, relocating the two logo PNGs, and hand-editing admin-front's copy of `AuthShell/index.tsx` (the only content edit in this task — everything else stays a byte-identical copy).

**Files:**
- Create (copy of `front/src/core/shared/ui/AuthShell/`, then edited — see Step 1b): `admin-front/src/core/shared/ui/AuthShell/`
- Create (copy of `front/src/core/shared/ui/FormSelect/`): `admin-front/src/core/shared/ui/FormSelect/`
- Create (copy of `front/src/core/shared/ui/Toast/`): `admin-front/src/core/shared/ui/Toast/`
- Create (copy of `front/src/core/shared/icons/IconKey.tsx`): `admin-front/src/core/shared/icons/IconKey.tsx`
- Create (copy of `front/src/core/shared/icons/IconChevronDown.tsx`): `admin-front/src/core/shared/icons/IconChevronDown.tsx`
- Create (copy of `front/src/core/shared/icons/types.ts`): `admin-front/src/core/shared/icons/types.ts`
- Create: `admin-front/src/core/shared/icons/index.ts` (new minimal barrel, not a copy — front's barrel exports 29 icons, admin-front only has 2)
- Create (copy of `front/src/core/shared/lib/formatPrice.ts`): `admin-front/src/core/shared/lib/formatPrice.ts`
- Create (copy of `front/src/core/shared/lib/readFormString.ts`): `admin-front/src/core/shared/lib/readFormString.ts`
- Create (copy of `front/src/core/shared/styles/`): `admin-front/src/core/shared/styles/`
- Create (copy of `front/src/core/shared/assets/fonts/`): `admin-front/src/core/shared/assets/fonts/`
- Create (copy of `front/src/core/shared/ui/Header/assets/logo.png` and `logo@2x.png`): `admin-front/src/core/shared/assets/logo.png`, `admin-front/src/core/shared/assets/logo@2x.png`

**Interfaces:**
- Produces: `@/core/shared/ui/{AuthShell,FormSelect,Toast}`, `@/core/shared/icons/{IconKey,IconChevronDown}` (direct paths) and `@/core/shared/icons` (barrel, for `FormSelect`'s import style), `@/core/shared/lib/{formatPrice,readFormString}`, `@/core/shared/styles/index.css` (imported by the root layout in Task 6).
- Consumes: `adminAuthErrorCleared` from `@/core/entities/adminSession` (Task 2). Forward-references `AppPath` from `@/core/shared/router/paths` — that file doesn't exist until Task 5 runs; this is expected and matches the same forward-dependency pattern Task 2 left for Task 3 (nothing in this task's own verification steps typechecks or builds, so it's fine for `AuthShell`/`AdminLoginPage`-style forward references to resolve later — Task 8 is where the full build proves everything ties together).

- [ ] **Step 1: Copy UI components**

```bash
mkdir -p admin-front/src/core/shared/ui admin-front/src/core/shared/icons admin-front/src/core/shared/lib
cp -r front/src/core/shared/ui/AuthShell admin-front/src/core/shared/ui/AuthShell
cp -r front/src/core/shared/ui/FormSelect admin-front/src/core/shared/ui/FormSelect
cp -r front/src/core/shared/ui/Toast admin-front/src/core/shared/ui/Toast
```

- [ ] **Step 1b: Edit admin-front's `AuthShell/index.tsx`**

Open `admin-front/src/core/shared/ui/AuthShell/index.tsx` (the copy just made — it is currently byte-identical to `front`'s) and make exactly these three changes:

1. Replace the import line:
```ts
import { authErrorCleared } from '@/core/entities/session';
```
with:
```ts
import { adminAuthErrorCleared } from '@/core/entities/adminSession';
```

2. Replace the two usages of `authErrorCleared` with `adminAuthErrorCleared` (the `useUnit(authErrorCleared)` call and nothing else — the local variable name `clearAuthError` stays as-is, it's still an accurate name):
```ts
const clearAuthError = useUnit(adminAuthErrorCleared);
```

3. Replace the two logo import lines:
```ts
import logoImage from '@/core/shared/ui/Header/assets/logo.png';
import logoImage2x from '@/core/shared/ui/Header/assets/logo@2x.png';
```
with:
```ts
import logoImage from '@/core/shared/assets/logo.png';
import logoImage2x from '@/core/shared/assets/logo@2x.png';
```

Leave everything else in the file — including both `AppPath.Home` links (the logo click-through and the "← На сайт" link) — untouched. `AppPath.Home` resolves to `/` in admin-front's own de-prefixed paths (Task 5), i.e. the admin dashboard: both links go to the admin home, not the public storefront. That's a deliberate, minimal scope call (admin-front has no public-site URL plumbed anywhere else either) — not a bug to fix in this task.

- [ ] **Step 2: No `core/shared/ui` barrel needed**

Confirmed: `front/src/core/shared/ui/index.ts` does not export `AuthShell`, `FormSelect`, or `Toast`/`ToastViewport` at all — every consumer (including `front/src/app/layout.tsx`, which does `import { ToastViewport } from '@/core/shared/ui/Toast';`) imports these directly by their own path, never through the barrel. `admin-front` needs no `core/shared/ui/index.ts` file — skip creating one.

- [ ] **Step 3: Copy icons, create the minimal barrel, and copy lib helpers**

```bash
cp front/src/core/shared/icons/IconKey.tsx admin-front/src/core/shared/icons/IconKey.tsx
cp front/src/core/shared/icons/IconChevronDown.tsx admin-front/src/core/shared/icons/IconChevronDown.tsx
cp front/src/core/shared/icons/types.ts admin-front/src/core/shared/icons/types.ts
cp front/src/core/shared/lib/formatPrice.ts admin-front/src/core/shared/lib/formatPrice.ts
cp front/src/core/shared/lib/readFormString.ts admin-front/src/core/shared/lib/readFormString.ts
```

Create `admin-front/src/core/shared/icons/index.ts` (new file, not a copy — front's barrel has 29 icon exports, admin-front only ships the 2 it uses):

```ts
export { IconChevronDown } from './IconChevronDown';
export { IconKey } from './IconKey';
export type { Icon, IconProps } from './types';
```

- [ ] **Step 4: No `core/shared/lib` barrel needed**

Confirmed: `front/src/core/shared/lib/index.ts` re-exports `breakpoints`, `useLayoutType`, and `validateContact` helpers — it does not export `formatPrice` or `readFormString` at all. `views/Admin` imports both directly (`from '@/core/shared/lib/formatPrice'`, `from '@/core/shared/lib/readFormString'`), never through the barrel. Skip creating a `lib/index.ts` in admin-front — it would only contain unused exports.

- [ ] **Step 5: Copy styles, fonts, and the logo images**

```bash
cp -r front/src/core/shared/styles admin-front/src/core/shared/styles
mkdir -p admin-front/src/core/shared/assets
cp -r front/src/core/shared/assets/fonts admin-front/src/core/shared/assets/fonts
cp front/src/core/shared/ui/Header/assets/logo.png admin-front/src/core/shared/assets/logo.png
cp front/src/core/shared/ui/Header/assets/logo@2x.png admin-front/src/core/shared/assets/logo@2x.png
```

- [ ] **Step 6: Verify copies**

Run: `diff -r front/src/core/shared/styles admin-front/src/core/shared/styles`
Expected: no output.

Run: `ls admin-front/src/core/shared/assets/fonts | wc -l`
Expected: `6` (same font files as front).

Run: `diff front/src/core/shared/ui/Header/assets/logo.png admin-front/src/core/shared/assets/logo.png && diff front/src/core/shared/ui/Header/assets/logo@2x.png admin-front/src/core/shared/assets/logo@2x.png`
Expected: no output (binary files identical).

Run: `diff front/src/core/shared/icons/IconChevronDown.tsx admin-front/src/core/shared/icons/IconChevronDown.tsx`
Expected: no output.

Run: `grep -n "authErrorCleared\|Header/assets" admin-front/src/core/shared/ui/AuthShell/index.tsx`
Expected: only `adminAuthErrorCleared` appears (twice), no `Header/assets` path.

- [ ] **Step 7: Commit**

```bash
git add admin-front/src/core/shared/ui admin-front/src/core/shared/icons admin-front/src/core/shared/lib admin-front/src/core/shared/styles admin-front/src/core/shared/assets
git commit -m "chore(admin-front): duplicate shared UI kit, icons, lib helpers, and styles from front"
```

---

## Task 5: Create admin-front's router paths and proxy middleware

De-prefixed paths (no more `/admin`) and a standalone auth guard that protects every route except `/login`.

**Files:**
- Create: `admin-front/src/core/shared/router/paths.ts`
- Create: `admin-front/src/proxy.ts`

**Interfaces:**
- Consumes: `ADMIN_USER_ID_COOKIE` from `@/core/entities/adminSession/lib/constants` (moved in Task 2).
- Produces: `AppPath` enum and `getContentPath`/`getProductPath` helpers that Task 6 rewrites `views/Admin`/`app` code to use instead of the old `AppPath.Admin*`/`getAdminContentPath`/`getAdminProductPath` names.

- [ ] **Step 1: Write `admin-front/src/core/shared/router/paths.ts`**

```ts
export enum AppPath {
  AuditLog = '/audit-log',
  Banners = '/banners',
  Categories = '/categories',
  Content = '/content',
  Feedback = '/feedback',
  Home = '/',
  Legal = '/legal',
  Login = '/login',
  Products = '/products',
  SignupRequests = '/signup-requests',
  Support = '/support',
  Users = '/users',
}

export const getContentPath = (pageKey: string): string => `${AppPath.Content}/${pageKey}`;

export const getProductPath = (productId: string): string => `${AppPath.Products}/${productId}`;
```

- [ ] **Step 2: Write `admin-front/src/proxy.ts`**

```ts
import type { NextRequest } from 'next/server';

import { NextResponse } from 'next/server';

import { ADMIN_USER_ID_COOKIE } from '@/core/entities/adminSession/lib/constants';
import { AppPath } from '@/core/shared/router/paths';

export const config = {
  matcher: ['/', '/((?!_next|favicon.ico).*)'],
};

export const proxy = (request: NextRequest): NextResponse => {
  const { pathname } = request.nextUrl;

  if (pathname === AppPath.Login) {
    return NextResponse.next();
  }

  const adminUserId = request.cookies.get(ADMIN_USER_ID_COOKIE)?.value;

  if (!adminUserId) {
    return NextResponse.redirect(new URL(AppPath.Login, request.url));
  }

  return NextResponse.next();
};
```

This is the same logic as the `/admin` branch of `front/src/proxy.ts` (Task 7 deletes that branch from `front`), minus the path-prefix check since every route here is an admin route.

- [ ] **Step 3: Commit**

```bash
git add admin-front/src/core/shared/router admin-front/src/proxy.ts
git commit -m "feat(admin-front): add de-prefixed router paths and auth proxy"
```

---

## Task 6: Move admin views and routes into admin-front, rewrite path references

**Correction found while preparing this task's dispatch:** `views/Admin` (`AdminDashboardPage.tsx`, `AdminBannersPage.tsx`, `AdminContentPage.tsx`, `AdminLegalPage.tsx`, and `views/Admin/model/content.ts`) imports `useContent`/`contentInvalidated` from `@/core/entities/content` — a small effector entity (store + `useContent` hook) that caches content/banners/legal docs, used by both the public site and admin. It only depends on `@/core/shared/api/content` and `@/core/shared/api/parseApiError` (both already duplicated in Task 3), so it's self-contained and safe to duplicate whole. The original file list omitted it entirely — added as Step 0 below.

**Files:**
- Create (copy of `front/src/core/entities/content/`): `admin-front/src/core/entities/content/`
- Move: `front/src/views/Admin/` → `admin-front/src/views/Admin/`
- Move: `front/src/app/admin/page.tsx` → `admin-front/src/app/page.tsx`
- Move: `front/src/app/admin/login/page.tsx` → `admin-front/src/app/login/page.tsx`
- Move: `front/src/app/admin/audit-log/page.tsx` → `admin-front/src/app/audit-log/page.tsx`
- Move: `front/src/app/admin/banners/page.tsx` → `admin-front/src/app/banners/page.tsx`
- Move: `front/src/app/admin/categories/page.tsx` → `admin-front/src/app/categories/page.tsx`
- Move: `front/src/app/admin/content/[pageKey]/page.tsx` → `admin-front/src/app/content/[pageKey]/page.tsx`
- Move: `front/src/app/admin/feedback/page.tsx` → `admin-front/src/app/feedback/page.tsx`
- Move: `front/src/app/admin/legal/page.tsx` → `admin-front/src/app/legal/page.tsx`
- Move: `front/src/app/admin/products/page.tsx` → `admin-front/src/app/products/page.tsx`
- Move: `front/src/app/admin/products/[productId]/page.tsx` → `admin-front/src/app/products/[productId]/page.tsx`
- Move: `front/src/app/admin/signup-requests/page.tsx` → `admin-front/src/app/signup-requests/page.tsx`
- Move: `front/src/app/admin/support/page.tsx` → `admin-front/src/app/support/page.tsx`
- Move: `front/src/app/admin/users/page.tsx` → `admin-front/src/app/users/page.tsx`
- Create: `admin-front/src/app/layout.tsx`
- Create: `admin-front/src/app/ReactScan.tsx` (copy of `front/src/app/ReactScan.tsx`)
- Modify: `admin-front/src/views/Admin/AdminDashboardPage.tsx`, `AdminLoginPage.tsx`, `AdminProductPage.tsx`, `AdminProductsPage.tsx`, `ui/AdminShell.tsx`, `lib/nav.ts` — rewrite `AppPath.Admin*` / `getAdminContentPath` / `getAdminProductPath` to the new names from Task 5.

**Interfaces:**
- Consumes: `AppPath`, `getContentPath`, `getProductPath` from `@/core/shared/router/paths` (Task 5); `useAdminSession`, `ADMIN_ACTOR_LABEL` from `@/core/entities/adminSession` (Task 2); `AuthShell`, `FormSelect`, `Toast/model`, `IconKey`, `formatPrice`, `readFormString` (Task 4); `content`/`feedback`/`support`/`products`/`admin`/`portalUsers`/`signupRequests` API clients (Tasks 2–3); `useContent`, `contentInvalidated` from `@/core/entities/content` (Step 0 below).
- Produces: full route tree of `admin-front` — `/`, `/login`, `/products`, `/products/:id`, `/categories`, `/content/:pageKey`, `/banners`, `/legal`, `/feedback`, `/support`, `/signup-requests`, `/users`, `/audit-log`.

- [ ] **Step 0: Duplicate the `content` entity**

```bash
mkdir -p admin-front/src/core/entities
cp -r front/src/core/entities/content admin-front/src/core/entities/content
diff -r front/src/core/entities/content admin-front/src/core/entities/content
```
Expected: `diff` produces no output (byte-identical copy).

- [ ] **Step 1: Move the views directory**

```bash
mkdir -p admin-front/src/views
git mv front/src/views/Admin admin-front/src/views/Admin
```

- [ ] **Step 2: Move route files into a de-prefixed tree**

```bash
mkdir -p admin-front/src/app/login admin-front/src/app/audit-log admin-front/src/app/banners \
  admin-front/src/app/categories "admin-front/src/app/content/[pageKey]" admin-front/src/app/feedback \
  admin-front/src/app/legal admin-front/src/app/products "admin-front/src/app/products/[productId]" \
  admin-front/src/app/signup-requests admin-front/src/app/support admin-front/src/app/users

git mv front/src/app/admin/page.tsx admin-front/src/app/page.tsx
git mv front/src/app/admin/login/page.tsx admin-front/src/app/login/page.tsx
git mv front/src/app/admin/audit-log/page.tsx admin-front/src/app/audit-log/page.tsx
git mv front/src/app/admin/banners/page.tsx admin-front/src/app/banners/page.tsx
git mv front/src/app/admin/categories/page.tsx admin-front/src/app/categories/page.tsx
git mv "front/src/app/admin/content/[pageKey]/page.tsx" "admin-front/src/app/content/[pageKey]/page.tsx"
git mv front/src/app/admin/feedback/page.tsx admin-front/src/app/feedback/page.tsx
git mv front/src/app/admin/legal/page.tsx admin-front/src/app/legal/page.tsx
git mv front/src/app/admin/products/page.tsx admin-front/src/app/products/page.tsx
git mv "front/src/app/admin/products/[productId]/page.tsx" "admin-front/src/app/products/[productId]/page.tsx"
git mv front/src/app/admin/signup-requests/page.tsx admin-front/src/app/signup-requests/page.tsx
git mv front/src/app/admin/support/page.tsx admin-front/src/app/support/page.tsx
git mv front/src/app/admin/users/page.tsx admin-front/src/app/users/page.tsx
```

None of these route files need content edits — they import from `@/views/Admin` via the barrel (`views/Admin/index.tsx`), which moved as-is in Step 1.

- [ ] **Step 3: Delete the now-empty `front/src/app/admin/layout.tsx`**

```bash
git rm front/src/app/admin/layout.tsx
```

Its logic (`<AdminShell>{children}</AdminShell>`) is folded into `admin-front/src/app/layout.tsx` in Step 5 — there's no other top-level route in admin-front that needs to be excluded from the admin chrome, so a separate nested layout isn't needed.

- [ ] **Step 4: Rewrite `AppPath.Admin*` references — apply to all six files below**

Use exact string replacement (`replace_all` where a file has more than one occurrence of the same old string), in this order per file — **specific identifiers first, bare `AppPath.Admin` / `AppPath.AdminLogin` last** — since `AppPath.Admin` is a substring of every other `AppPath.AdminXxx` identifier and must not be replaced until those are gone.

Global mapping used throughout:

| Old | New |
|---|---|
| `AppPath.AdminAuditLog` | `AppPath.AuditLog` |
| `AppPath.AdminBanners` | `AppPath.Banners` |
| `AppPath.AdminCategories` | `AppPath.Categories` |
| `AppPath.AdminContent` | `AppPath.Content` |
| `AppPath.AdminFeedback` | `AppPath.Feedback` |
| `AppPath.AdminLegal` | `AppPath.Legal` |
| `AppPath.AdminProducts` | `AppPath.Products` |
| `AppPath.AdminSignupRequests` | `AppPath.SignupRequests` |
| `AppPath.AdminSupport` | `AppPath.Support` |
| `AppPath.AdminUsers` | `AppPath.Users` |
| `AppPath.AdminLogin` | `AppPath.Login` |
| `AppPath.Admin` (bare, after all the above are done) | `AppPath.Home` |
| `getAdminContentPath` | `getContentPath` |
| `getAdminProductPath` | `getProductPath` |

`admin-front/src/views/Admin/AdminDashboardPage.tsx`:
- `import { AppPath, getAdminContentPath } from '@/core/shared/router/paths';` → `import { AppPath, getContentPath } from '@/core/shared/router/paths';`
- `href={AppPath.AdminLegal}` → `href={AppPath.Legal}`
- `href={AppPath.AdminAuditLog}` → `href={AppPath.AuditLog}`
- `href={AppPath.AdminBanners}` → `href={AppPath.Banners}`
- `href={AppPath.AdminProducts}` → `href={AppPath.Products}`
- `href={AppPath.AdminSignupRequests}` → `href={AppPath.SignupRequests}`
- `href={AppPath.AdminSupport}` → `href={AppPath.Support}`
- `href={getAdminContentPath('home')}` → `href={getContentPath('home')}`

`admin-front/src/views/Admin/AdminLoginPage.tsx`:
- `router.replace(AppPath.Admin);` → `router.replace(AppPath.Home);` (`replace_all`, occurs twice with identical surrounding code)

`admin-front/src/views/Admin/AdminProductPage.tsx`:
- `AppPath.AdminProducts` → `AppPath.Products` (`replace_all`, occurs twice: `router.push(AppPath.AdminProducts)` and `href={AppPath.AdminProducts}`)

`admin-front/src/views/Admin/AdminProductsPage.tsx`:
- `import { getAdminProductPath } from '@/core/shared/router/paths';` → `import { getProductPath } from '@/core/shared/router/paths';`
- `getAdminProductPath` → `getProductPath` (`replace_all`, occurs 3 times: `getAdminProductPath('new')`, and twice `getAdminProductPath(product.id)`)

`admin-front/src/views/Admin/ui/AdminShell.tsx`:
- `AppPath.AdminLogin` → `AppPath.Login` (`replace_all`, occurs 3 times)
- `AppPath.Admin` → `AppPath.Home` (`replace_all`, occurs 3 times: `href={AppPath.Admin}`, `item.href === AppPath.Admin`, `pathname === AppPath.Admin`) — do this **after** the `AdminLogin` replacement above so it doesn't also mangle those.

`admin-front/src/views/Admin/lib/nav.ts`:
- `import { AppPath, getAdminContentPath } from '@/core/shared/router/paths';` → `import { AppPath, getContentPath } from '@/core/shared/router/paths';`
- `AppPath.Admin` → `AppPath.Home` (single occurrence: `{ href: AppPath.Admin, label: 'Сводка' }`)
- `getAdminContentPath` → `getContentPath` (`replace_all`, occurs 6 times, one per content page key)
- `AppPath.AdminBanners` → `AppPath.Banners`
- `AppPath.AdminProducts` → `AppPath.Products`
- `AppPath.AdminCategories` → `AppPath.Categories`
- `AppPath.AdminLegal` → `AppPath.Legal`
- `AppPath.AdminUsers` → `AppPath.Users`
- `AppPath.AdminSignupRequests` → `AppPath.SignupRequests`
- `AppPath.AdminFeedback` → `AppPath.Feedback`
- `AppPath.AdminSupport` → `AppPath.Support`
- `AppPath.AdminAuditLog` → `AppPath.AuditLog`

- [ ] **Step 5: Write `admin-front/src/app/layout.tsx`**

Merges the old root layout (`front/src/app/layout.tsx`) and the old nested admin layout (`AdminShell` wrapper, deleted in Step 3) into one:

```tsx
import type { Metadata } from 'next';

import { ReactScan } from '@/app/ReactScan';
import '@/core/shared/styles/index.css';
import { AdminShell } from '@/views/Admin';
import { ToastViewport } from '@/core/shared/ui/Toast';

export const metadata: Metadata = {
  title: 'AMC Admin',
};

type RootLayoutProps = {
  children: React.ReactNode;
};

const RootLayout = ({ children }: RootLayoutProps): JSX.Element => {
  return (
    <html lang="ru">
      <ReactScan />
      <body>
        <AdminShell>{children}</AdminShell>
        <ToastViewport />
      </body>
    </html>
  );
};

export default RootLayout;
```

`AdminShell` is exported from `views/Admin/index.tsx`'s barrel (confirmed during design) and `ToastViewport` from `core/shared/ui/Toast/index.tsx` directly — this matches exactly how `front/src/app/layout.tsx` imports it today (`import { ToastViewport } from '@/core/shared/ui/Toast';`), confirmed neither goes through a `core/shared/ui` barrel (see Task 4 Step 2).

- [ ] **Step 6: Copy `ReactScan.tsx`**

```bash
cp front/src/app/ReactScan.tsx admin-front/src/app/ReactScan.tsx
```

- [ ] **Step 7: Verify no `AppPath.Admin` or `getAdmin*Path` references remain in admin-front**

Run: `grep -rn "AppPath\.Admin\|getAdminContentPath\|getAdminProductPath" admin-front/src`
Expected: no output.

- [ ] **Step 8: Commit**

```bash
git add -A
git commit -m "feat(admin-front): move admin views and routes, de-prefix paths"
```

---

## Task 7: Remove admin code from front and verify it still builds

**Files:**
- Delete: `front/src/app/admin/` (now empty except already-removed files — remove the whole directory)
- Modify: `front/src/proxy.ts` — remove the `/admin` branch
- Modify: `front/src/core/shared/router/paths.ts` — remove all `Admin*` enum members and `getAdminContentPath`/`getAdminProductPath`

**Interfaces:**
- Consumes: nothing new.
- Produces: a `front` that builds, typechecks, and lints clean with zero references to the moved admin code.

- [ ] **Step 1: Remove the leftover admin route directory**

```bash
find front/src/app/admin -type f
```

Expected: no output (Task 6 moved every file out and Task 6 Step 3 removed `layout.tsx`). If anything is still listed, `git mv`/`git rm` it into place per Task 6 before continuing.

```bash
rmdir front/src/app/admin 2>/dev/null || true
```

- [ ] **Step 2: Edit `front/src/proxy.ts`**

Current content:

```ts
import type { NextRequest } from 'next/server';

import { NextResponse } from 'next/server';

import { ADMIN_USER_ID_COOKIE } from '@/core/entities/adminSession/lib/constants';
import { USER_ID_COOKIE } from '@/core/entities/session/lib/constants';
import { AppPath } from '@/core/shared/router/paths';

export const config = {
  matcher: ['/admin', '/admin/:path*', '/cabinet', '/cabinet/:path*', '/checkout'],
};

export const proxy = (request: NextRequest): NextResponse => {
  const { pathname } = request.nextUrl;

  if (pathname.startsWith(AppPath.Admin)) {
    if (pathname === AppPath.AdminLogin) {
      return NextResponse.next();
    }

    const adminUserId = request.cookies.get(ADMIN_USER_ID_COOKIE)?.value;

    if (!adminUserId) {
      return NextResponse.redirect(new URL(AppPath.AdminLogin, request.url));
    }

    return NextResponse.next();
  }

  const userId = request.cookies.get(USER_ID_COOKIE)?.value;

  if (!userId) {
    const loginUrl = new URL(AppPath.Login, request.url);

    loginUrl.searchParams.set('next', pathname);

    return NextResponse.redirect(loginUrl);
  }

  return NextResponse.next();
};
```

Replace with:

```ts
import type { NextRequest } from 'next/server';

import { NextResponse } from 'next/server';

import { USER_ID_COOKIE } from '@/core/entities/session/lib/constants';
import { AppPath } from '@/core/shared/router/paths';

export const config = {
  matcher: ['/cabinet', '/cabinet/:path*', '/checkout'],
};

export const proxy = (request: NextRequest): NextResponse => {
  const { pathname } = request.nextUrl;

  const userId = request.cookies.get(USER_ID_COOKIE)?.value;

  if (!userId) {
    const loginUrl = new URL(AppPath.Login, request.url);

    loginUrl.searchParams.set('next', pathname);

    return NextResponse.redirect(loginUrl);
  }

  return NextResponse.next();
};
```

- [ ] **Step 3: Edit `front/src/core/shared/router/paths.ts`**

Remove these lines from the `AppPath` enum:

```ts
  Admin = '/admin',
  AdminAuditLog = '/admin/audit-log',
  AdminBanners = '/admin/banners',
  AdminCategories = '/admin/categories',
  AdminContent = '/admin/content',
  AdminFeedback = '/admin/feedback',
  AdminLegal = '/admin/legal',
  AdminLogin = '/admin/login',
  AdminProducts = '/admin/products',
  AdminSignupRequests = '/admin/signup-requests',
  AdminSupport = '/admin/support',
  AdminUsers = '/admin/users',
```

Remove these two functions entirely:

```ts
export const getAdminContentPath = (pageKey: string): string =>
  `${AppPath.AdminContent}/${pageKey}`;

export const getAdminProductPath = (productId: string): string =>
  `${AppPath.AdminProducts}/${productId}`;
```

- [ ] **Step 4: Verify no dangling references anywhere in front**

Run: `grep -rln "AppPath\.Admin\|adminSession\|getAdminContentPath\|getAdminProductPath\|views/Admin\|core/shared/api/admin'\|core/shared/api/portalUsers'\|core/shared/api/signupRequests'" front/src`
Expected: no output.

- [ ] **Step 5: Typecheck, lint, and build front**

Run: `cd front && yarn typecheck`
Expected: exits 0, no errors.

Run: `cd front && yarn lint`
Expected: exits 0, no errors.

Run: `cd front && yarn build`
Expected: build succeeds, no route still lists `/admin/*` in the build output route table.

- [ ] **Step 6: Commit**

```bash
git add -A
git commit -m "refactor(front): remove admin panel code, now served by admin-front"
```

---

## Task 8: Verify admin-front builds standalone

**Environment note:** the system's default `node` is v21.4.0, incompatible with this project's eslint config (needs 20.19+/22.13+/24+). A compatible `node@22` is installed via Homebrew at `/Users/macbook/.brew/opt/node@22/bin`, not on default PATH. Prepend it for every yarn command: `export PATH="/Users/macbook/.brew/opt/node@22/bin:$PATH"`.

**Note on unexpected prior commits:** at the time this task was prepared, three commits (each titled `fix new problem`, same author as the rest of this branch) had already landed on top of Task 7's commit, made by a process outside this plan's execution. They: (1) added `admin-front/Dockerfile`, a `deploy/docker-compose.yml` `admin-front` service, and a CI matrix entry in `.github/workflows/build.yml` — with **different values than Task 9 below specifies** (port 3001 not 3000, a direct `ADMIN_API_URL` env pointing at the internal `admin` backend service instead of the nginx+`/portal-api/`-proxy design, and a Dockerfile that still provisions an unused `/data` volume for `PORTAL_DATA_FILE`); (2) added a missing `@react-aria/utils` entry to `admin-front/yarn.lock`; (3) added `admin-front/src/types/jsx.d.ts` (a global `JSX.Element` shim) and reordered two import lines in `admin-front/src/app/layout.tsx`. Do not revert any of this — treat it as already-landed groundwork. Task 9 (next) reconciles the deploy-wiring discrepancies; this task only needs to get `admin-front` typechecking/linting/building clean, and the `jsx.d.ts` shim already covers one class of typecheck error.

**Diagnosis already run once before this task's dispatch (findings below), so the fixes are known — this task executes them, not discovers them from scratch:**

**Files:**
- Modify: `admin-front/src/core/shared/router/paths.ts` — add one exported constant.
- Modify: `admin-front/src/views/Admin/AdminCategoriesPage.tsx` — one line, use the new constant.
- Modify: `admin-front/src/views/Admin/AdminDashboardPage.tsx` — reorder two import lines (auto-fixable).
- Modify: `admin-front/src/views/Admin/AdminProductsPage.tsx` — reorder two named imports (auto-fixable).
- Modify: `admin-front/.env.example` — document the new env var.

**Interfaces:**
- Consumes: everything from Tasks 1–6, plus the three externally-landed commits described above.
- Produces: confidence that `admin-front` is a fully working, independently buildable Next.js app.

- [ ] **Step 1: Install and typecheck**

```bash
export PATH="/Users/macbook/.brew/opt/node@22/bin:$PATH"
cd admin-front && yarn install && yarn typecheck
```

Expected failure (already diagnosed): exactly one error —
```
src/views/Admin/AdminCategoriesPage.tsx(45,34): error TS2339: Property 'Catalog' does not exist on type 'typeof AppPath'.
```
`AdminCategoriesPage.tsx` links to the **public storefront's** catalog page filtered by category (`` `${AppPath.Catalog}?categoryID=${node.id}` `` in the original `front` code) — a legitimate "preview this category on the live site" link, not an admin-internal route. `admin-front`'s de-prefixed `AppPath` enum has no public-site paths at all (by design — Task 5 only gave it admin routes). Fix by adding a new exported constant to `admin-front/src/core/shared/router/paths.ts` (append after the `AppPath` enum, don't add `Catalog` to the enum itself — the enum is for admin-front's own routes only):

```ts
export const PUBLIC_CATALOG_URL = `${process.env.NEXT_PUBLIC_SITE_ORIGIN ?? 'https://wk.amctechgroup.ru'}/catalog`;
```

In `admin-front/src/views/Admin/AdminCategoriesPage.tsx`, change:
```tsx
<Link href={`${AppPath.Catalog}?categoryID=${node.id}`}>{node.name}</Link>
```
to:
```tsx
<Link href={`${PUBLIC_CATALOG_URL}?categoryID=${node.id}`}>{node.name}</Link>
```
Confirmed: `AppPath` is imported only for this one `.Catalog` usage in this file (line 10: `import { AppPath } from '@/core/shared/router/paths';`, no other `AppPath.` reference anywhere else in the file). Replace that import line with:
```ts
import { PUBLIC_CATALOG_URL } from '@/core/shared/router/paths';
```

Append to `admin-front/.env.example` (the file already exists from Task 1 — just add these two lines at the end, don't recreate it):
```
# Публичный домен фронта — используется только для ссылок «превью на сайте»
# (например, категория товаров из AdminCategoriesPage)
NEXT_PUBLIC_SITE_ORIGIN=https://wk.amctechgroup.ru
```

Re-run `yarn typecheck` after the fix. Expected: exits 0.

- [ ] **Step 2: Lint**

```bash
yarn lint
```

Expected (already diagnosed): two trivial, auto-fixable import-order errors, plus a set of pre-existing issues inherited unchanged from `front`'s original admin code (same code, same lint config, same rule severities — these are not new regressions introduced by the split; see the file:line list below).

Fix the two auto-fixable ones — confirm each is genuinely just import ordering (not a real bug) before fixing:
- `admin-front/src/views/Admin/AdminDashboardPage.tsx:14` — `perfectionist/sort-imports` (`./model/audit` should come before `./model/catalog`)
- `admin-front/src/views/Admin/AdminProductsPage.tsx:15` — `perfectionist/sort-named-imports` (`$adminProductQuery` before `$adminProducts`)

```bash
yarn lint:js:fix
```

**Leave the following as documented pre-existing debt inherited from `front` (do not attempt to fix — the underlying application logic and effector patterns are unchanged by this migration, and fixing them risks behavior changes out of this task's scope; matches the precedent Task 7 set for `front`'s own pre-existing lint debt):**
- `react-hooks/set-state-in-effect` (Important-looking but pre-existing) in `AdminLegalPage.tsx:32`, `AdminLegalPage.tsx:40`, `AdminProductPage.tsx:77`
- `effector/no-watch` warning in `AdminProductPage.tsx:116`
- `Unused eslint-disable directive` warnings in `views/Admin/model/content.ts:58`, `model/feedback.ts:74`, `model/users.ts:76`
- `@typescript-eslint/no-unsafe-enum-comparison` errors in `ui/AdminShell.tsx:26,76,77` (same `pathname === AppPath.X` pattern existed in `front`'s original file before Task 6's rename — verified via `git show` against the pre-Task-6 commit; the identifier names changed but the type-level shape did not)

Re-run `yarn lint` after the two auto-fixes. Expected: the count drops by exactly the two fixed errors (from 17 errors/5 warnings to 15 errors/5 warnings); the remaining errors are all in the "leave as-is" list above. If `yarn lint` exits non-zero because ESLint's exit code reflects any remaining error (it will, since errors — not just warnings — remain), that's expected and fine: this task's bar is "no NEW errors beyond the documented pre-existing list", not "exit code 0". Note this explicitly in your report so the task reviewer doesn't mistake it for an unaddressed failure.

- [ ] **Step 3: Build**

```bash
yarn build
```

Expected: build succeeds (Next.js's `next build` does not run ESLint by default in this project — only `yarn lint` does — so Step 2's remaining pre-existing lint errors do not block this). Confirm the printed route table matches: `/`, `/login`, `/products`, `/products/[productId]`, `/categories`, `/content/[pageKey]`, `/banners`, `/legal`, `/feedback`, `/support`, `/signup-requests`, `/users`, `/audit-log`.

- [ ] **Step 4: Manual smoke test (best-effort)**

```bash
cd admin-front && API_PROXY_TARGET=https://wk.amctechgroup.ru PORTAL_API_PROXY_TARGET=https://wk.amctechgroup.ru NEXT_PUBLIC_SITE_ORIGIN=https://wk.amctechgroup.ru yarn dev
```
Open `http://localhost:3000/` in a browser if one is available in this environment — expect a redirect to `/login` (no `admin_user_id` cookie yet). If the sandbox has no network access to `wk.amctechgroup.ru` or no browser, skip actually logging in — just confirm the dev server starts without crashing and the root path redirects to `/login` (checkable via `curl -sI http://localhost:3000/` — expect a 307/308 redirect to `/login`). Stop the dev server after confirming (Ctrl-C).

- [ ] **Step 5: Commit fixes**

```bash
git add admin-front/src/core/shared/router/paths.ts admin-front/src/views/Admin/AdminCategoriesPage.tsx admin-front/src/views/Admin/AdminDashboardPage.tsx admin-front/src/views/Admin/AdminProductsPage.tsx admin-front/.env.example
git commit -m "fix(admin-front): resolve build/typecheck issues found during verification"
```

---

## Task 9: Wire up deployment — Dockerfile, docker-compose, nginx, CI, husky

**Correction found while preparing this task's dispatch — read this before starting:** while Task 7/8 were running, three commits from a process outside this plan's execution (confirmed not the human operator — asked directly) landed on this branch and already partially did this task's work, with different, partly-broken choices. Current state, verified directly against the checkout:

- `admin-front/Dockerfile` **already exists** — multi-stage build matching the right pattern, but: (a) port is `3001` not `3000` (harmless, arbitrary, no code cares — **this plan now adopts 3001** to avoid needless churn against already-landed work, so every `3000` below is intentionally `3001` instead of what earlier tasks in this plan file said); (b) still provisions a `/data` directory and a comment about `PORTAL_DATA_FILE` — dead weight, `admin-front` never reads that env var or writes to that path (only `front` does), must be removed.
- `deploy/docker-compose.yml` **already has** an `admin-front` service block, but: (a) indentation is inconsistent with its sibling services (6-space `environment:` block vs. the file's 4-space convention throughout) — cosmetic, fix for consistency; (b) sets `ADMIN_API_URL: "http://admin:8084"`, which **no code anywhere in `admin-front` reads** (confirmed via repo-wide grep — dead config, likely an abandoned attempt at a different design) — must be replaced with the env vars `admin-front`'s actual `next.config.ts` (Task 1) reads: `API_PROXY_TARGET`, `PORTAL_API_PROXY_TARGET`, plus `NEXT_PUBLIC_SITE_ORIGIN` (Task 8, for the public-catalog preview link); (c) `depends_on` only lists `admin`, should be `access` and `front` (matching this plan's original design — `front` because of the `/portal-api/` proxy dependency).
- `.github/workflows/build.yml` **already has** the correct `admin-front` matrix entry (dockerfile path and image name exactly match what this task would have written) — no change needed, skip that part of Step 4.
- `deploy/nginx/conf.d/admin.wk.amctechgroup.ru.conf` **does not exist** — nobody added it. Without it, `admin-front` isn't reachable on its own domain at all regardless of the other wiring. Still needed in full, per Step 3 below (with `admin-front:3001` instead of `:3000`).
- `.husky/pre-commit` **is untouched** — Step 5 below still applies as originally planned.

**Files:**
- Modify (not create — already exists): `admin-front/Dockerfile` — remove the dead `/data`/`PORTAL_DATA_FILE` provisioning, keep port `3001`.
- Modify (not create — already exists): `deploy/docker-compose.yml` — fix indentation, replace `ADMIN_API_URL` with the real env vars, fix `depends_on`.
- Create: `deploy/nginx/conf.d/admin.wk.amctechgroup.ru.conf`
- Skip: `.github/workflows/build.yml` — already correct, verify only, don't edit.
- Modify: `.husky/pre-commit` — run lint-staged in both apps

**Interfaces:** none (infrastructure only).

- [ ] **Step 1: Fix `admin-front/Dockerfile`**

Read the current file first (`cat admin-front/Dockerfile`) to confirm it still matches what's described above before editing — if it's already changed since this brief was written, stop and report rather than guessing. Remove exactly this block (the comment plus the `RUN mkdir` line) from the `runner` stage, and nothing else — keep `EXPOSE 3001` / `ENV PORT=3001` as they are:

```dockerfile
# Каталог для PORTAL_DATA_FILE. Named volume наследует владельца этой
# директории при первом монтировании — иначе процесс под юзером app
# не смог бы писать состояние портальных модулей.
RUN mkdir -p /data && chown app:app /data

```
(Delete this whole block, including the blank line that follows it before `USER app` — leave exactly one blank line between the `COPY --from=builder ... .next/static` line and `USER app`, matching `front/Dockerfile`'s spacing convention.)

The rest of the file (deps/builder/runner stages, `COPY` lines, `ENTRYPOINT`) stays exactly as-is.

- [ ] **Step 2: Fix the `admin-front` service in `deploy/docker-compose.yml`**

Read the current block first (`grep -A 12 "^  admin-front:" deploy/docker-compose.yml`) to confirm it still matches what's described above. Replace the entire `admin-front:` service block with:

```yaml
  admin-front:
    image: ghcr.io/mbatimel/amc-admin-front:${IMAGE_TAG}
    restart: unless-stopped
    environment:
      PORT: "3001"
      HOSTNAME: "0.0.0.0"
      API_PROXY_TARGET: "https://wk.amctechgroup.ru"
      PORTAL_API_PROXY_TARGET: "https://wk.amctechgroup.ru"
      NEXT_PUBLIC_SITE_ORIGIN: "https://wk.amctechgroup.ru"
    depends_on:
      - access
      - front
    networks: [amc_net]
```

(4-space indentation throughout, matching every other service in this file — the existing block's `environment:` sub-keys are indented 6 spaces, that's the inconsistency being fixed.)

And add `admin-front` to the `nginx` service's `depends_on` list (alongside the existing `access`, `admin`, `auth`, `orders`, `products`, `users`, `front`, `swagger-ui`) — check first whether it's already there (the external commits didn't add it, but verify).

- [ ] **Step 3: Write `deploy/nginx/conf.d/admin.wk.amctechgroup.ru.conf`**

```nginx
# TLS терминирует внешний слой перед хостом, сюда трафик приходит уже как
# обычный HTTP через хостовый nginx -> 127.0.0.1:8090.

server {
    listen 80;
    server_name admin.wk.amctechgroup.ru;

    location /api/v1/access/ {
        proxy_pass http://access:8080/api/v1/access/;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
    location /api/v1/auth/ {
        proxy_pass http://auth:8081/api/v1/auth/;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
    location /api/v1/admin/ {
        proxy_pass http://admin:8084/api/v1/admin/;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
    location /api/v1/products {
        proxy_pass http://products:8085/api/v1/products;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
    location /api/v1/categories {
        proxy_pass http://products:8085/api/v1/categories;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
    location /api/v1/brands {
        proxy_pass http://products:8085/api/v1/brands;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
    location /api/v1/users/ {
        proxy_pass http://users:8083/api/v1/users/;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    location /portal-api/ {
        proxy_pass http://front:3000/portal-api/;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    location / {
        proxy_pass http://admin-front:3001;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

`/api/v1/orders` is intentionally omitted — the admin panel has no order-management screens (confirmed: `AdminShell`/`nav.ts` cover audit-log, banners, categories, content, feedback, legal, products, signup-requests, support, users only). `/portal-api/` is proxied straight to the `front` container (port `3000`, unaffected by `admin-front`'s port) so `admin-front`'s browser-side fetches (relative `/portal-api/...` calls from `portalClient.ts`) hit the single source of truth without needing CORS. `admin-front` itself listens on `3001` (see Step 1's correction) — note the two different ports, don't "fix" the `3001` here to `3000`, that would break the compose service.

- [ ] **Step 4: Verify the Docker image build matrix (already correct, no edit needed)**

Run: `grep -A2 "name: admin-front" .github/workflows/build.yml`
Expected:
```yaml
          - name: admin-front
            dockerfile: ./admin-front/Dockerfile
            image: ghcr.io/mbatimel/amc-admin-front
```
This was already added by the external commits and matches exactly what this task would have written — confirm it, don't touch the file.

- [ ] **Step 5: Update `.husky/pre-commit` to lint-stage both apps**

Current content: `cd front && yarn lint-staged`

Replace with:

```bash
cd front && yarn lint-staged
cd ../admin-front && yarn lint-staged
```

- [ ] **Step 6: Verify docker-compose config parses (best-effort)**

Run: `cd deploy && docker compose config --quiet`
Expected: exits 0, no YAML/schema errors. If the `docker` CLI isn't available in this environment, skip this step and note it in the report rather than blocking on it — it's a syntax/schema check, not something the rest of this task depends on. (This validates syntax only — it does not require `IMAGE_TAG`/`PG_*` to resolve to real images, `--quiet` suppresses the rendered output but still validates.)

- [ ] **Step 7: Commit**

```bash
git add admin-front/Dockerfile deploy/docker-compose.yml deploy/nginx/conf.d/admin.wk.amctechgroup.ru.conf .husky/pre-commit
git commit -m "feat(deploy): fix admin-front Dockerfile/compose wiring, add nginx domain and husky lint-staged"
```

`.github/workflows/build.yml` is deliberately excluded — Step 4 confirmed it needs no changes.

---

## Post-plan follow-ups (not part of this plan, worth a mention to the user)

- `deploy/.env.example` may want a note that `admin.wk.amctechgroup.ru` needs its own DNS record and TLS cert at the external edge (same as the existing `wk.amctechgroup.ru`/`amctechgroup.ru` setup) — outside this repo's scope.
- Once `back/admin` grows real endpoints for content/banners/legal/support/feedback, `front/src/app/portal-api` and `front/src/core/shared/server/portal` get deleted per that directory's own README — at that point the `/portal-api/` nginx location and the `PORTAL_API_PROXY_TARGET` rewrite in `admin-front/next.config.ts` become dead and should be removed too.
