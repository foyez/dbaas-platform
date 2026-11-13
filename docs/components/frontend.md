# Web UI

The UI (`client/`) is a Vue 3 + TypeScript single-page app, built with Vite and styled with Tailwind CSS. It covers the whole instance lifecycle: create, watch status, view credentials and logs, run SQL, update and delete. A public landing page and product docs sit in front of the authenticated app.

## Pages

| Route | Page |
|---|---|
| `/` | Landing page |
| `/docs/*` | Product docs for platform users: getting started, architecture, API, SDK |
| `/dashboard` | Overview of the user's instances |
| `/instances`, `/instances/new`, `/instances/:id` | List, create and detail pages; the detail page has the SQL console |
| `/audit-logs` | Audit trail, visible to `dbaas.admin` only |

## How it works

- **Login** uses the Authorization Code flow with PKCE through `oidc-client-ts` and `vue-oidc-context`. The access token is sent as a `Bearer` token on every API call.
- **Live status** comes from polling. `usePolling` uses a recursive `setTimeout` instead of `setInterval`, so a slow response never causes overlapping requests.
- **Forms** are validated with `zod` before anything is sent to the API.
- **State** lives in Pinia stores.

## Configuration

Vite bakes these into the bundle at build time, so each environment gets its own image:

| Variable | Purpose |
|---|---|
| `VITE_API_URL` | Base URL of the REST API |
| `VITE_ZITADEL_AUTHORITY` | ZITADEL issuer URL |
| `VITE_ZITADEL_CLIENT_ID` | Client ID of the `dbaas-web` OIDC app |
| `VITE_ZITADEL_PROJECT_ID` | Project ID, used to request the API audience |
| `VITE_ZITADEL_REDIRECT_URI` | Where ZITADEL sends the user after login |
| `VITE_ZITADEL_POST_LOGOUT_URI` | Where ZITADEL sends the user after logout |

The `Containerfile` refuses to build if any of them is missing.

## Develop and test

```sh
cd client
npm install
npm run dev          # Vite dev server
npm run test:unit    # Vitest
npm run test:e2e     # Playwright
```
