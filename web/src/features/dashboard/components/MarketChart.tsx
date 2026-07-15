import { AreaChart } from "@mantine/charts";
import { Badge, Group, Paper, SegmentedControl, Skeleton, Text } from "@mantine/core";
import { EmptyState } from "../../../components/feedback/EmptyState";
import { formatMoney } from "../../../lib/formatters";
import type { ChartPeriod, ChartPoint, Quote } from "../types";
import { PriceChange } from "./PriceChange";

interface MarketChartProps {
  symbol: string;
  quote?: Quote;
  data: ChartPoint[];
  loading: boolean;
  period: ChartPeriod;
  onPeriodChange: (period: ChartPeriod) => void;
}

const periods: ChartPeriod[] = ["1D", "1W", "1M", "3M", "1Y"];

export function MarketChart({
  symbol,
  quote,
  data,
  loading,
  period,
  onPeriodChange,
}: MarketChartProps) {
  return (
    <Paper className="panel chart-panel" p={{ base: "md", sm: "xl" }}>
      <Group justify="space-between" align="flex-start" mb="lg">
        <div>
          <Group gap="xs">
            <Text fw={650} fz="lg">
              {symbol}
            </Text>
            <Badge variant="light" color="gray" size="sm">
              {quote?.currency || "Currency unavailable"}
            </Badge>
          </Group>
          <Group gap="sm" mt={6} align="baseline">
            <Text fz="xl" fw={650}>
              {quote ? formatMoney(quote.price, quote.currency) : "—"}
            </Text>
            {quote && <PriceChange value={quote.priceChangePercent} />}
          </Group>
        </div>
        <SegmentedControl
          size="xs"
          value={period}
          onChange={(value) => onPeriodChange(value as ChartPeriod)}
          data={periods}
        />
      </Group>

      {loading ? (
        <Skeleton h={310} radius="md" />
      ) : data.length ? (
        <AreaChart
          h={310}
          data={data}
          dataKey="date"
          series={[{ name: "price", color: "lime.5" }]}
          curveType="monotone"
          withDots={false}
          withLegend={false}
          gridAxis="none"
          yAxisProps={{
            domain: ["auto", "auto"],
            tickFormatter: (value) => formatMoney(Number(value), quote?.currency),
          }}
          tooltipProps={{
            content: ({ label, payload }) => <ChartTooltip label={label} payload={payload} currency={quote?.currency} />,
          }}
        />
      ) : (
        <EmptyState
          title="No chart data"
          detail="Historical prices will appear when the stock service responds."
        />
      )}
    </Paper>
  );
}

function ChartTooltip({
  label,
  payload,
  currency,
}: {
  label?: React.ReactNode;
  payload?: readonly { value?: unknown }[];
  currency?: string;
}) {
  return (
    <Paper withBorder shadow="md" p="sm">
      <Text size="xs" c="dimmed">
        {label}
      </Text>
      <Text fw={650}>
        {payload?.[0]?.value !== undefined ? formatMoney(Number(payload[0].value), currency) : "—"}
      </Text>
    </Paper>
  );
}
