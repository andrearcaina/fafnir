import { Badge, Button, Group, Paper, SimpleGrid, Stack, Text, Title } from "@mantine/core";
import { IconArrowLeft, IconShoppingCart, IconStar, IconStarFilled } from "@tabler/icons-react";
import { formatCompactNumber, formatMoney } from "../../../lib/formatters";
import { MarketChart } from "../../dashboard/components/MarketChart";
import { MetricCard } from "../../dashboard/components/MetricCard";
import type { ChartPeriod, ChartPoint, Quote } from "../../dashboard/types";
import type { StockMetadata } from "../../../types/domain";
import { useWatchlist } from "../api/useWatchlist";

interface StockDetailPageProps {
  symbol: string;
  quote?: Quote;
  metadata?: StockMetadata;
  chartData: ChartPoint[];
  chartLoading: boolean;
  period: ChartPeriod;
  onPeriodChange: (period: ChartPeriod) => void;
  onBack: () => void;
  onTrade: () => void;
  isWatchlisted: boolean;
}

export function StockDetailPage({
  symbol,
  quote,
  metadata,
  chartData,
  chartLoading,
  period,
  onPeriodChange,
  onBack,
  onTrade,
  isWatchlisted,
}: StockDetailPageProps) {
  const watchlist = useWatchlist(symbol, isWatchlisted);
  const instrumentType = metadata?.instrumentType.toUpperCase();
  const isTradable = symbol.length <= 10 && (instrumentType === "EQUITY" || instrumentType === "ETF");

  return (
    <Stack gap="lg">
      <Group justify="space-between" align="flex-end">
        <div>
          <Button
            variant="subtle"
            color="gray"
            size="compact-sm"
            px={0}
            mb="md"
            leftSection={<IconArrowLeft size={16} />}
            onClick={onBack}
          >
            Back to markets
          </Button>
          <Group gap="sm">
            <Title order={1} className="page-title">{symbol}</Title>
            <Badge color="gray" variant="light">{metadata?.exchangeFullName || metadata?.exchange || "Exchange unavailable"}</Badge>
          </Group>
          <Text c="dimmed" size="sm" mt={4}>{metadata?.name ?? "Stock detail and market performance"}</Text>
        </div>
        <Group gap="sm">
          <Button
            variant={isWatchlisted ? "light" : "default"}
            color={isWatchlisted ? "yellow" : undefined}
            leftSection={isWatchlisted ? <IconStarFilled size={17} /> : <IconStar size={17} />}
            loading={watchlist.isPending}
            onClick={watchlist.toggle}
          >
            {isWatchlisted ? "Remove from watchlist" : "Add to watchlist"}
          </Button>
          <Button leftSection={<IconShoppingCart size={17} />} onClick={onTrade} disabled={!isTradable}>
            {isTradable ? `Trade ${symbol}` : "Market data only"}
          </Button>
        </Group>
      </Group>

      <MarketChart
        symbol={symbol}
        quote={quote}
        data={chartData}
        loading={chartLoading}
        period={period}
        onPeriodChange={onPeriodChange}
      />

      <SimpleGrid cols={{ base: 2, md: 4 }} spacing="md">
        <MetricCard label="Open" value={quote ? formatMoney(quote.open, quote.currency) : "—"} detail="Today's open" loading={!quote} />
        <MetricCard label="Previous close" value={quote ? formatMoney(quote.previousClose, quote.currency) : "—"} detail="Last session" loading={!quote} />
        <MetricCard label="Day range" value={quote ? `${formatMoney(quote.dayLow, quote.currency)} – ${formatMoney(quote.dayHigh, quote.currency)}` : "—"} detail="Low to high" loading={!quote} />
        <MetricCard label="Volume" value={quote ? formatCompactNumber(quote.volume) : "—"} detail={quote ? `Market cap ${formatCompactNumber(quote.marketCap)}` : "Market activity"} loading={!quote} />
      </SimpleGrid>

      <Paper className="panel" p="lg">
        <Group justify="space-between" align="flex-start">
          <div>
            <Text fw={650}>Market data</Text>
            <Text c="dimmed" size="sm" mt={4}>
              {quote?.asOf ? `As of ${new Date(quote.asOf).toLocaleString("en-CA")}` : "Timestamp unavailable"}
            </Text>
          </div>
          <Group gap="xs">
            {quote?.marketState && <Badge variant="light">{quote.marketState}</Badge>}
            <Badge color="gray" variant="light">{quote?.source || "Unknown source"}</Badge>
            {metadata?.currency && <Badge color="gray" variant="outline">{metadata.currency}</Badge>}
            {metadata?.instrumentType && <Badge color="gray" variant="outline">{metadata.instrumentType}</Badge>}
          </Group>
        </Group>
      </Paper>

      <Paper className="panel" p="lg">
        <Text fw={650}>52-week range</Text>
        <Group justify="space-between" mt="md">
          <div><Text size="xs" c="dimmed">Low</Text><Text fw={600}>{quote ? formatMoney(quote.yearLow, quote.currency) : "—"}</Text></div>
          <div style={{ textAlign: "right" }}><Text size="xs" c="dimmed">High</Text><Text fw={600}>{quote ? formatMoney(quote.yearHigh, quote.currency) : "—"}</Text></div>
        </Group>
      </Paper>
    </Stack>
  );
}
