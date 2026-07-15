import { useMemo, useState } from "react";
import { ActionIcon, Group, Menu, Paper, Skeleton, Stack, Text, UnstyledButton } from "@mantine/core";
import { IconCheck, IconMenu2 } from "@tabler/icons-react";
import { formatCompactNumber, formatMoney } from "../../../lib/formatters";
import type { Quote } from "../types";
import { PriceChange } from "./PriceChange";

interface MarketListProps {
  quotes: Quote[];
  activeSymbol: string;
  loading: boolean;
  onSelect: (symbol: string) => void;
}

export function MarketList({ quotes, activeSymbol, loading, onSelect }: MarketListProps) {
  const [sort, setSort] = useState<"symbol" | "change" | "price">("symbol");
  const sortedQuotes = useMemo(
    () =>
      [...quotes].sort((a, b) => {
        if (sort === "change") return b.priceChangePercent - a.priceChangePercent;
        if (sort === "price") return b.price - a.price;
        return a.symbol.localeCompare(b.symbol);
      }),
    [quotes, sort],
  );

  return (
    <Paper className="panel watch-panel" p="lg">
      <Group justify="space-between" mb="md">
        <Text fw={650}>Markets</Text>
        <Menu position="bottom-end" width={180}>
          <Menu.Target>
            <ActionIcon variant="subtle" color="gray" aria-label="Sort markets">
              <IconMenu2 size={17} />
            </ActionIcon>
          </Menu.Target>
          <Menu.Dropdown>
            <Menu.Label>Sort markets</Menu.Label>
            <SortItem label="Symbol" active={sort === "symbol"} onClick={() => setSort("symbol")} />
            <SortItem label="Top gainers" active={sort === "change"} onClick={() => setSort("change")} />
            <SortItem label="Highest price" active={sort === "price"} onClick={() => setSort("price")} />
          </Menu.Dropdown>
        </Menu>
      </Group>
      <Stack gap={2}>
        {loading
          ? Array.from({ length: 5 }).map((_, index) => <Skeleton key={index} h={58} />)
          : sortedQuotes.map((quote) => (
              <UnstyledButton
                key={quote.symbol}
                className="quote-row"
                data-active={quote.symbol === activeSymbol || undefined}
                onClick={() => onSelect(quote.symbol)}
              >
                <Group justify="space-between" wrap="nowrap">
                  <div>
                    <Text fw={650} size="sm">
                      {quote.symbol}
                    </Text>
                    <Text size="xs" c="dimmed">
                      Vol {formatCompactNumber(quote.volume)}
                    </Text>
                  </div>
                  <div className="quote-price">
                    <Text fw={600} size="sm">
                      {formatMoney(quote.price, quote.currency)}
                    </Text>
                    <PriceChange value={quote.priceChangePercent} compact />
                  </div>
                </Group>
              </UnstyledButton>
            ))}
      </Stack>
    </Paper>
  );
}

function SortItem({
  label,
  active,
  onClick,
}: {
  label: string;
  active: boolean;
  onClick: () => void;
}) {
  return (
    <Menu.Item onClick={onClick} rightSection={active ? <IconCheck size={14} /> : undefined}>
      {label}
    </Menu.Item>
  );
}
