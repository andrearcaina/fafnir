import type { DashboardQuery, MarketQuotesQuery } from "../gql/graphql";

export type Profile = NonNullable<DashboardQuery["getProfileData"]["data"]>;
export type Quote = NonNullable<
  NonNullable<MarketQuotesQuery["getStockQuoteBatch"]["data"]>[number]
>;
export type Account = NonNullable<
  NonNullable<DashboardQuery["getPortfolioSummary"]["accounts"]>[number]
>;
export type Order = NonNullable<NonNullable<DashboardQuery["getOrders"]["data"]>[number]>;
