import { Button, Group, Paper, SimpleGrid, Stack, Text } from "@mantine/core";
import { formatAccountBalances } from "../../../lib/formatters";
import { OrdersTable } from "../../orders/components/OrdersTable";
import { MarketChart } from "../components/MarketChart";
import { MarketList } from "../components/MarketList";
import { MetricCard } from "../components/MetricCard";
import type { Account, ChartPeriod, ChartPoint, Order, Quote } from "../types";

interface OverviewSectionProps {
  loading: boolean;
  accounts: Account[];
  quotes: Quote[];
  orders: Order[];
  activeSymbol: string;
  activeQuote?: Quote;
  chartData: ChartPoint[];
  chartLoading: boolean;
  period: ChartPeriod;
  onSymbolChange: (symbol: string) => void;
  onPeriodChange: (period: ChartPeriod) => void;
  onTrade: () => void;
}

export function OverviewSection({
  loading,
  accounts,
  quotes,
  orders,
  activeSymbol,
  activeQuote,
  chartData,
  chartLoading,
  period,
  onSymbolChange,
  onPeriodChange,
  onTrade,
}: OverviewSectionProps) {
  const accountBalances = formatAccountBalances(accounts);
  const openOrders = orders.filter(
    (order) => !["FILLED", "CANCELED", "REJECTED"].includes(order.status),
  ).length;

  return (
    <Stack gap="lg">
      <SimpleGrid cols={{ base: 1, sm: 3 }} spacing="md">
        <MetricCard
          label="Total balance"
          value={accountBalances}
          detail={`${accounts.length} account${accounts.length === 1 ? "" : "s"}`}
          loading={loading}
          featured
        />
        <MetricCard
          label="Buying power"
          value={accountBalances}
          detail="Available to trade"
          loading={loading}
        />
        <MetricCard
          label="Open orders"
          value={String(openOrders)}
          detail={`${orders.length} total orders`}
          loading={loading}
        />
      </SimpleGrid>

      <SimpleGrid cols={{ base: 1, lg: 3 }} spacing="lg" className="overview-grid">
        <MarketChart
          symbol={activeSymbol}
          quote={activeQuote}
          data={chartData}
          loading={chartLoading}
          period={period}
          onPeriodChange={onPeriodChange}
        />
        <MarketList
          quotes={quotes}
          activeSymbol={activeSymbol}
          loading={loading}
          onSelect={onSymbolChange}
        />
      </SimpleGrid>

      <Paper className="panel" p={{ base: "md", sm: "lg" }}>
        <Group justify="space-between" mb="md">
          <div>
            <Text fw={650}>Recent orders</Text>
            <Text c="dimmed" size="xs" mt={2}>
              Your latest market activity
            </Text>
          </div>
          <Button variant="subtle" size="xs" onClick={onTrade}>
            New order
          </Button>
        </Group>
        <OrdersTable orders={orders.slice(0, 5)} />
      </Paper>
    </Stack>
  );
}
