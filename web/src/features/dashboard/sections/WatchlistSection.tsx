import { Avatar, Center, Group, Paper, SimpleGrid, Stack, Text, UnstyledButton } from "@mantine/core";
import { IconStar } from "@tabler/icons-react";
import { formatCompactNumber, formatMoney } from "../../../lib/formatters";
import { PriceChange } from "../components/PriceChange";
import type { Quote } from "../types";

interface WatchlistSectionProps {
  quotes: Quote[];
  savedSymbols: string[];
  onSelect: (symbol: string) => void;
}

export function WatchlistSection({
  quotes,
  savedSymbols,
  onSelect,
}: WatchlistSectionProps) {
  const shownQuotes = quotes.filter((quote) => savedSymbols.includes(quote.symbol));

  if (!savedSymbols.length) {
    return (
      <Paper className="panel" p="xl">
        <Center mih={240}>
          <Stack align="center" gap="xs">
            <IconStar size={32} stroke={1.5} />
            <Text fw={650}>Your watchlist is empty</Text>
            <Text c="dimmed" size="sm" ta="center">
              Search for a stock, open its page, then select Add to watchlist.
            </Text>
          </Stack>
        </Center>
      </Paper>
    );
  }

  return (
    <SimpleGrid cols={{ base: 1, sm: 2, xl: 3 }}>
      {shownQuotes.map((quote) => (
        <UnstyledButton key={quote.symbol} onClick={() => onSelect(quote.symbol)}>
          <Paper className="panel quote-card" p="lg">
            <Group justify="space-between">
              <Avatar color="dark" radius="xl">
                {quote.symbol.slice(0, 1)}
              </Avatar>
              <PriceChange value={quote.priceChangePercent} />
            </Group>
            <Text fw={700} mt="lg">
              {quote.symbol}
            </Text>
            <Group justify="space-between" mt={4}>
              <Text fz="xl" fw={650}>
                {formatMoney(quote.price)}
              </Text>
              <Text c="dimmed" size="xs">
                Vol {formatCompactNumber(quote.volume)}
              </Text>
            </Group>
          </Paper>
        </UnstyledButton>
      ))}
    </SimpleGrid>
  );
}
