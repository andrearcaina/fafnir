import type { Account, Order, Profile, Quote, StockMetadata } from "../../types/domain";

export type { Account, Order, Profile, Quote, StockMetadata } from "../../types/domain";

export type DashboardSection = "Overview" | "Portfolio" | "Orders" | "Watchlist";
export type ChartPeriod = "1D" | "1W" | "1M" | "3M" | "1Y";

export interface ChartPoint {
  date: string;
  price: number;
}

export interface DashboardSnapshot {
  profile?: Profile;
  accounts: Account[];
  quotes: Quote[];
  marketQuotes: Quote[];
  orders: Order[];
  watchlistSymbols: string[];
  activeQuote?: Quote;
  activeMetadata?: StockMetadata;
  chartData: ChartPoint[];
  isLoading: boolean;
  isChartLoading: boolean;
}
