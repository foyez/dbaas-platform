# Frontend: Vue 3 SPA

!!! abstract "Where it's used in this project"
    The web UI in [`client/`](https://github.com/foyez/dbaas-platform/tree/main/client): Vue 3, TypeScript, Vite, Tailwind and shadcn-vue, served by unprivileged nginx. See [Web UI](../components/frontend.md).

## Key ideas

### Vue 3 for a React developer

| Vue | React equivalent | Note |
|---|---|---|
| Single-file component (`.vue`) | Component file | Template, script and style in one file |
| `<script setup>` + Composition API | Function component + hooks | Runs **once** per component instance, not on every render |
| `ref()` / `reactive()` | `useState` | Mutate directly (`count.value++`); Vue tracks dependencies |
| `computed()` | `useMemo` | Dependencies are tracked automatically |
| `watch()` / `watchEffect()` | `useEffect` | For side effects in response to changes |
| Composable (`useX()`) | Custom hook | Plain function returning reactive state |
| `v-model` | Controlled input | Two-way binding, also on custom components |
| Slots | `children` / render props | Named and scoped slots |
| `provide` / `inject` | Context | Dependency passing without prop drilling |
| Pinia | Zustand / Redux | Global stores |

The key mental shift: Vue's reactivity tracks what you read, so there are no dependency arrays and no re-running of the whole component function.

### Serving a SPA in a container

- **Multi-stage build**: Node builds static files; the runtime image is nginx with only `dist/`.
- **`nginx-unprivileged`** runs as a non-root user on a high port, so it works under a restrictive security context.
- **`try_files $uri $uri/ /index.html`** sends deep links such as `/instances/abc` to the SPA, so Vue Router can handle them instead of nginx returning 404.
- **Cache hashed assets forever, but not `index.html`.** Files under `/assets/` have content hashes in their names, so they can be `immutable` for a year; `index.html` must stay fresh so a new deploy is picked up.
- **`VITE_*` variables are baked in at build time.** The image is per environment, unlike the API, which reads config at runtime.

## How this project uses it

- **Feature composables** (`useInstances`, `useInstance`, `useCreateInstance`) hold the data fetching; pages only render.
- **Zod schemas** validate forms before sending, for user feedback. The API still validates everything, because the client can't be trusted.
- **The create form reuses one idempotency key** for retries of the same submission and only generates a new one after success, so a double click or retry can't create two databases.
- **Polling** uses recursive `setTimeout` so requests never overlap, stops when an instance reaches a terminal status, and stops when the component unmounts.

## What I learned building it

- **Generate the idempotency key per operation, not per request.** A new key on every retry defeats the point.
- **Polling needs exit conditions**: terminal status, unmount and ideally a maximum duration. Otherwise a broken backend keeps a browser polling forever.
- **After a write, use the response.** One bug kept showing stale data after saving an update until the page applied the update's response and restarted polling.
- **Small API client bugs hide easily.** `updateInstance` once called the credentials endpoint; a typed API module plus unit tests (`api/instances.test.ts`) catch this kind of mistake.
- **Be explicit about `index.html` caching.** The nginx config marks `/assets/` immutable but sets no header on `index.html`, which leaves it to browser heuristics; `Cache-Control: no-cache` there would make deploys reliable.
- **Don't build a dashboard endpoint too early.** The dashboard reuses the list endpoint until there's a real need for aggregated data.

## Review questions

??? question "Why does `<script setup>` not need dependency arrays like React hooks?"
    It runs once per component instance. Reactivity is tracked at runtime: when a `computed` or `watch` reads a `ref`, Vue records that dependency and re-runs only that piece when it changes.

??? question "Why does the nginx config need `try_files … /index.html`?"
    Routes like `/instances/123` exist only in the client-side router. Without the fallback, nginx looks for a file at that path and returns 404 on refresh or deep links.

??? question "If the UI validates with Zod, why does the API validate again?"
    Client-side validation is for user experience. Anyone can call the API directly, so the server must remain the authority for correctness and security.

??? question "Why use recursive `setTimeout` instead of `setInterval` for polling?"
    `setInterval` fires on schedule even if the previous request hasn't finished, so slow responses pile up. Recursive `setTimeout` schedules the next tick only after the current one completes.

??? question "Why are frontend environment variables a build-time concern?"
    Vite replaces `import.meta.env.VITE_*` with literal values when bundling, so the built JavaScript already contains them. Changing them means rebuilding the image.
