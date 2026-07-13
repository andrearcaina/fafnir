import type { DashboardSection } from "../types";
import type { ChartPeriod, DashboardSnapshot } from "../types";
import { OrdersSection } from "./OrdersSection";
import { OverviewSection } from "./OverviewSection";
import { PortfolioSection } from "./PortfolioSection";
import { WatchlistSection } from "./WatchlistSection";

interface DashboardContentProps {
  section: DashboardSection;
  data: DashboardSnapshot;
  activeSymbol: string;
  period: ChartPeriod;
  onSymbolChange: (symbol: string) => void;
  onPeriodChange: (period: ChartPeriod) => void;
  onWatchlistSelect: (symbol: string) => void;
  onTrade: () => void;
}

export function DashboardContent({
  section,
  data,
  activeSymbol,
  period,
  onSymbolChange,
  onPeriodChange,
  onWatchlistSelect,
  onTrade,
}: DashboardContentProps) {
  switch (section) {
    case "Portfolio":
      return (
        <PortfolioSection
          totalBalance={data.totalBalance}
          accounts={data.accounts}
          loading={data.isLoading}
        />
      );
    case "Orders":
      return <OrdersSection orders={data.orders} />;
    case "Watchlist":
      return (
        <WatchlistSection
          quotes={data.quotes}
          savedSymbols={data.watchlistSymbols}
          onSelect={onWatchlistSelect}
        />
      );
    case "Overview":
      return (
        <OverviewSection
          loading={data.isLoading}
          totalBalance={data.totalBalance}
          accounts={data.accounts}
          quotes={data.marketQuotes}
          orders={data.orders}
          activeSymbol={activeSymbol}
          activeQuote={data.activeQuote}
          chartData={data.chartData}
          chartLoading={data.isChartLoading}
          period={period}
          onSymbolChange={onSymbolChange}
          onPeriodChange={onPeriodChange}
          onTrade={onTrade}
        />
      );
  }
}
