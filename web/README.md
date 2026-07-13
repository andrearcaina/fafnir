# Fafnir web

React + TypeScript frontend for the Fafnir API gateway. It uses Mantine for UI, TanStack Query for server state, `graphql-request` for transport, and GraphQL Code Generator for schema-derived types.

## Local development

Start the backend gateway on port `8080`, then:

```bash
cd web
npm install
npm run dev
```

Vite proxies `/auth` and `/graphql` to the gateway so cookie authentication works without local CORS configuration. Override the target with `VITE_DEV_GATEWAY_URL` if needed.

## Commands

- `npm run dev` — start the Vite development server
- `npm run codegen` — generate TypeScript types from the local gqlgen schemas
- `npm run typecheck` — regenerate types and run TypeScript checks
- `npm run build` — create a production build

In production, serve the built SPA and gateway under the same origin, or configure credentialed CORS explicitly.

## Source organization

The frontend is organized by product feature rather than file type:

```text
src/
├── app/                    # providers and global theme
├── components/             # feature-agnostic UI
├── features/
│   ├── auth/               # auth API hooks, form, and page
│   ├── dashboard/          # dashboard data, layout, sections, and market UI
│   ├── orders/             # reusable order presentation
│   └── trading/            # order mutation and ticket
├── gql/                    # generated types (not committed)
├── lib/                    # API clients and pure shared utilities
└── styles/                 # global-only CSS
```

Feature components do not call GraphQL directly. Operations, requests, and cache behavior live in each feature's `api/` directory, while shared domain types are derived from generated GraphQL output.
