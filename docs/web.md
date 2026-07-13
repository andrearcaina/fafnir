# Web Architecture

The Fafnir frontend is a React and TypeScript application inside [`web/`](../web).
Its design is inspired by Wealthsimple and TradingView: dark, clean, and focused
on market data.

## Tech Stack

| Technology | Use |
| --- | --- |
| React and TypeScript | Application and components |
| Vite | Development and production builds |
| Mantine | UI components, forms, theme, and notifications |
| TanStack Query | Server data and mutations |
| GraphQL Request | API requests |
| React Router | Page navigation |
| Mantine Charts and Recharts | Stock charts |
| Tabler Icons | Icons |

## Structure

```text
web/src/
├── app/             # Providers and theme
├── components/      # Shared UI
├── features/        # Product features
├── gql/             # Generated API types
├── lib/             # API and shared utilities
├── styles/          # Global styles
├── App.tsx          # Main application
└── main.tsx         # Browser entry point
```

The frontend uses a feature-first structure. Each feature owns its API calls,
components, hooks, and pages.

```text
features/stocks/
├── api/
├── components/
└── pages/
```

Only reusable, feature-independent components belong in `src/components`.
Generated files in `src/gql` should not be edited manually.

## Application Layout

Global providers live in `src/app`. They configure Mantine, TanStack Query,
notifications, and routing.

`App.tsx` loads either the login screen or the authenticated dashboard. The
dashboard coordinates navigation and dialogs, while feature folders contain
the actual screens and logic.

## Routes

| Route | Screen |
| --- | --- |
| `/` | Dashboard, portfolio, orders, and watchlist |
| `/stocks/:symbol` | Stock chart, details, trading, and watchlist controls |
| `/settings` | Profile and account settings |
