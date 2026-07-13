import { useMemo } from "react";
import { useQuery } from "@tanstack/react-query";
import { DashboardDocument, MarketQuotesDocument, StockHistoryDocument } from "../../../gql/graphql";
import { graphQLClient } from "../../../lib/api";
import { formatShortDate } from "../../../lib/formatters";
import { isPresent } from "../../../lib/predicates";
import type { ChartPeriod } from "../types";

const DEFAULT_MARKET_SYMBOLS = ["AAPL", "MSFT", "NVDA", "TSLA", "AMZN"];
const MARKET_REFRESH_INTERVAL = 30_000;

export function useDashboardData(activeSymbol: string, period: ChartPeriod) {
  const dashboardQuery = useQuery({
    queryKey: ["dashboard"],
    queryFn: () => graphQLClient.request(DashboardDocument),
  });

  const watchlistSymbols = (dashboardQuery.data?.getWatchlist.data ?? [])
    .filter(isPresent)
    .map((item) => item.symbol);
  const quoteSymbols = Array.from(
    new Set([...DEFAULT_MARKET_SYMBOLS, activeSymbol, ...watchlistSymbols]),
  );

  const quotesQuery = useQuery({
    queryKey: ["market-quotes", quoteSymbols],
    queryFn: () => graphQLClient.request(MarketQuotesDocument, { symbols: quoteSymbols }),
    refetchInterval: MARKET_REFRESH_INTERVAL,
  });

  const historyQuery = useQuery({
    queryKey: ["stock-history", activeSymbol, period],
    queryFn: () =>
      graphQLClient.request(StockHistoryDocument, { symbol: activeSymbol, period }),
  });

  const data = dashboardQuery.data;
  const quotes = (quotesQuery.data?.getStockQuoteBatch.data ?? []).filter(isPresent);
  const orders = (data?.getOrders.data ?? []).filter(isPresent);
  const accounts = (data?.getPortfolioSummary.accounts ?? []).filter(isPresent);

  const chartData = useMemo(
    () =>
      (historyQuery.data?.getStockHistoricalData.data ?? [])
        .filter(isPresent)
        .map((point) => ({ date: formatShortDate(point.date), price: point.close })),
    [historyQuery.data],
  );

  return {
    profile: data?.getProfileData.data ?? undefined,
    totalBalance: data?.getPortfolioSummary.totalBalance ?? 0,
    accounts,
    quotes,
    marketQuotes: quotes.filter((quote) => DEFAULT_MARKET_SYMBOLS.includes(quote.symbol)),
    orders,
    watchlistSymbols,
    activeQuote: quotes.find((quote) => quote.symbol === activeSymbol) ?? quotes[0],
    chartData,
    isLoading: dashboardQuery.isPending || quotesQuery.isPending,
    isChartLoading: historyQuery.isPending,
    isRefreshing: dashboardQuery.isFetching || quotesQuery.isFetching,
    error: dashboardQuery.error ?? quotesQuery.error,
    refresh: () => Promise.all([dashboardQuery.refetch(), quotesQuery.refetch()]),
  };
}
