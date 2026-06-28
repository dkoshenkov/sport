# Project Instructions

## Structure

- `client/` contains the React app.
- `server/` is reserved for backend work, including future S3-backed exercise GIF access.
- Keep frontend, backend, auth, and exercise-media concerns separated.

## Product Context

- This is a focused workout program tool, not a marketing site or generic SaaS dashboard.
- The app is expected to convert spreadsheet-like training program data into a compact, readable interface.
- The service is account-gated: users without an account must not pass into the app.
- MVP auth is local: unique `nickname` plus password.
- Store only a password hash; never store or return raw passwords.
- Authenticated API requests use an HttpOnly `sid` session cookie.
- Expose a JS-readable cookie named `init` with value `1` when the app/session is initialized.
- Do not treat `init=1` as real authentication or authorization. It is only a temporary bootstrap marker.
- Store athlete defaults in profile: main 1RM values, preferred variant, preferred progression step, and optional notes.
- Store active program settings as a cycle snapshot, so profile changes do not silently rewrite in-progress or completed cycles.
- Default progression step is `4%` for XLSX compatibility; `5%` is allowed only as an explicit profile/cycle setting.

## Stack

- React
- TypeScript
- Tailwind CSS
- Vite

## Frontend Direction

- Build a focused training tool, not a landing page.
- The interface should feel like a polished spreadsheet-to-app conversion.
- Prioritize readability, density, forms, tables, and exercise details.
- Use a restrained visual system.
- Prefer small local components over a heavy UI kit.

## Allowed

- Tailwind utility classes.
- Small local components.
- Native inputs and selects when enough.
- `shadcn/ui` only for primitive components if explicitly useful: Button, Input, Select, Dialog, Table, Separator.

## Avoid

- Generic AI dashboard UI.
- Hero sections.
- Marketing layouts.
- Fake analytics.
- Fake charts.
- Bento grids.
- Glassmorphism.
- Gradient backgrounds.
- Cyberpunk gym aesthetics.
- Excessive cards.
- Emoji.
- Stock illustrations.
- Profile menus.
- Gamification.
- Calories, streaks, achievements, or fake metrics unless explicitly required.

## Component Rules

- Keep business logic outside React components.
- Keep XLSX-derived calculations in domain functions.
- Keep exercise dataset mapping in a separate alias file.
- Lazy-load GIF media.
- Handle missing GIFs with a useful empty state.
- Preserve mobile usability.
- Keep form, table, and exercise-detail interactions keyboard usable.

## Exercise Media

- The planned exercise source is `hasaneyldrm/exercises-dataset`.
- Do not hard-code remote dataset paths throughout components.
- Introduce a mapping layer for exercise aliases, normalized names, and GIF locations.
- Treat S3 as a server/storage concern. The client should consume stable URLs or API responses, not S3 implementation details.
- GIF previews must reserve dimensions to avoid layout shift.

## Skill Workflow

Use the global skills in this order when the task involves UI design or implementation:

1. `frontend-design`
2. `dembrandt`, plus focused dembrandt skills such as `visual-emphasis-and-hierarchy`, `modular-scale-typography`, `form-design`, `data-display-and-selection`, `ui-density`, and `responsive-paradigms`
3. Implementation
4. `baseline-ui`
5. `web-design-guidelines`
6. `accessibility-general`, `forms`, `keyboard`, and `tables` when relevant
7. `web-quality-audit` and `core-web-vitals`

Before implementing UI:

1. Describe screen structure.
2. Describe component tree.
3. Define visual direction.
4. Define Tailwind token/classes strategy.
5. List interaction states.

After implementing UI:

1. Run a baseline UI review.
2. Run a web design guidelines review.
3. Run an accessibility review.
4. Run a web quality review, especially for GIF loading, layout shift, input responsiveness, and bundle weight.
