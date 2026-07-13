import { Badge, Group, Paper, SimpleGrid, Stack, Text, ThemeIcon } from "@mantine/core";
import { IconWallet } from "@tabler/icons-react";
import { EmptyState } from "../../../components/feedback/EmptyState";
import { formatMoney } from "../../../lib/formatters";
import { MetricCard } from "../components/MetricCard";
import type { Account } from "../types";

interface PortfolioSectionProps {
  totalBalance: number;
  accounts: Account[];
  loading: boolean;
}

export function PortfolioSection({
  totalBalance,
  accounts,
  loading,
}: PortfolioSectionProps) {
  return (
    <Stack gap="lg">
      <MetricCard
        label="Net deposits"
        value={formatMoney(totalBalance)}
        detail="Across all currencies"
        loading={loading}
        featured
      />
      <SimpleGrid cols={{ base: 1, sm: 2, xl: 3 }}>
        {accounts.map((account) => (
          <Paper className="panel account-card" p="xl" key={account.id}>
            <Group justify="space-between">
              <ThemeIcon color="lime" variant="light" size="lg">
                <IconWallet size={20} />
              </ThemeIcon>
              <Badge color="gray" variant="light">
                {account.currency}
              </Badge>
            </Group>
            <Text tt="capitalize" c="dimmed" size="sm" mt="xl">
              {account.type.toLowerCase()}
            </Text>
            <Text fz={28} fw={650} mt={3}>
              {formatMoney(account.balance)}
            </Text>
            <Text c="dimmed" size="xs" mt="sm">
              •••• {account.accountNumber.slice(-4)}
            </Text>
          </Paper>
        ))}
      </SimpleGrid>
      {!accounts.length && !loading && (
        <EmptyState
          title="No accounts yet"
          detail="Create an account through the portfolio API to see it here."
        />
      )}
    </Stack>
  );
}
