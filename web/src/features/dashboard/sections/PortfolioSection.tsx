import { useEffect, useRef, useState } from "react";
import {
  ActionIcon,
  Badge,
  Button,
  Group,
  Modal,
  Paper,
  ScrollArea,
  SimpleGrid,
  Skeleton,
  Stack,
  Table,
  Tabs,
  Text,
  ThemeIcon,
  Tooltip,
} from "@mantine/core";
import { IconArrowsExchange, IconEye, IconTrash, IconWallet } from "@tabler/icons-react";
import { EmptyState } from "../../../components/feedback/EmptyState";
import { formatAccountBalances, formatDate, formatMoney, toTitleCase } from "../../../lib/formatters";
import { isPresent } from "../../../lib/predicates";
import { useDeletePortfolioAccount, useAccountActivity, useHoldingDetails } from "../../portfolio/api/usePortfolio";
import { TransferDialog } from "../../portfolio/components/TransferDialog";
import { MetricCard } from "../components/MetricCard";
import type { Account } from "../types";

interface PortfolioSectionProps {
  accounts: Account[];
  loading: boolean;
}

export function PortfolioSection({ accounts, loading }: PortfolioSectionProps) {
  const accountDetailsRef = useRef<HTMLDivElement>(null);
  const [selectedAccountId, setSelectedAccountId] = useState<string>();
  const [selectedSymbol, setSelectedSymbol] = useState<string>();
  const [transferOpened, setTransferOpened] = useState(false);
  const [deleteOpened, setDeleteOpened] = useState(false);
  const selectedAccount = accounts.find((account) => account.id === selectedAccountId) ?? accounts[0];
  const activity = useAccountActivity(selectedAccount?.id);
  const holdingDetails = useHoldingDetails(selectedAccount?.id, selectedSymbol);
  const deleteAccount = useDeletePortfolioAccount(() => {
    setDeleteOpened(false);
    setSelectedAccountId(undefined);
  });
  const holdings = (activity.data?.getHoldings.data ?? []).filter(isPresent);
  const transactions = (activity.data?.getTransactions.data ?? []).filter(isPresent);

  const selectAccount = (accountId: string) => {
    setSelectedAccountId(accountId);
    setSelectedSymbol(undefined);
    requestAnimationFrame(() => {
      accountDetailsRef.current?.scrollIntoView({ behavior: "smooth", block: "start" });
    });
  };

  useEffect(() => {
    if (!accounts.some((account) => account.id === selectedAccountId)) {
      setSelectedAccountId(accounts[0]?.id);
    }
  }, [accounts, selectedAccountId]);

  return (
    <Stack gap="lg">
      <Group justify="space-between" align="flex-end">
        <MetricCard
          label="Net deposits"
          value={formatAccountBalances(accounts)}
          detail="Grouped by account currency"
          loading={loading}
          featured
        />
        <Button
          variant="light"
          leftSection={<IconArrowsExchange size={17} />}
          disabled={accounts.length < 2}
          onClick={() => setTransferOpened(true)}
        >
          Transfer
        </Button>
      </Group>

      <SimpleGrid cols={{ base: 1, sm: 2, xl: 3 }}>
        {accounts.map((account) => (
          <Paper
            className="panel account-card"
            p="xl"
            key={account.id}
            withBorder={account.id === selectedAccount?.id}
            onClick={() => selectAccount(account.id)}
            onKeyDown={(event) => {
              if (event.key === "Enter" || event.key === " ") {
                event.preventDefault();
                selectAccount(account.id);
              }
            }}
            role="button"
            tabIndex={0}
            aria-pressed={account.id === selectedAccount?.id}
            aria-label={`View ${account.currency} ${account.type.toLowerCase()} account`}
            style={{ cursor: "pointer" }}
          >
            <Group justify="space-between">
              <ThemeIcon color="lime" variant="light" size="lg">
                <IconWallet size={20} />
              </ThemeIcon>
              <Badge color="gray" variant="light">{account.currency}</Badge>
            </Group>
            <Text tt="capitalize" c="dimmed" size="sm" mt="xl">{account.type.toLowerCase()}</Text>
            <Text fz={28} fw={650} mt={3}>{formatMoney(account.balance, account.currency)}</Text>
            <Text c="dimmed" size="xs" mt="sm">•••• {account.accountNumber.slice(-4)}</Text>
          </Paper>
        ))}
      </SimpleGrid>

      {!accounts.length && !loading ? (
        <EmptyState title="No accounts yet" detail="Open an account to view holdings and activity." />
      ) : selectedAccount ? (
        <Paper ref={accountDetailsRef} className="panel" p={{ base: "md", sm: "xl" }}>
          <Group justify="space-between" mb="lg">
            <div>
              <Text fw={650} fz="lg">{toTitleCase(selectedAccount.type)} account</Text>
              <Text c="dimmed" size="sm">Holdings and transaction history for •••• {selectedAccount.accountNumber.slice(-4)}</Text>
            </div>
            <Tooltip label="Delete this account">
              <ActionIcon color="red" variant="light" onClick={() => setDeleteOpened(true)} aria-label="Delete account">
                <IconTrash size={17} />
              </ActionIcon>
            </Tooltip>
          </Group>

          <Tabs defaultValue="holdings">
            <Tabs.List>
              <Tabs.Tab value="holdings">Holdings ({holdings.length})</Tabs.Tab>
              <Tabs.Tab value="transactions">Transactions ({transactions.length})</Tabs.Tab>
            </Tabs.List>

            <Tabs.Panel value="holdings" pt="lg">
              {activity.isPending ? <Skeleton h={140} /> : activity.isError ? (
                <Text c="red.4" size="sm">Could not load account holdings.</Text>
              ) : holdings.length ? (
                <ScrollArea>
                  <Table verticalSpacing="sm" miw={620}>
                    <Table.Thead><Table.Tr><Table.Th>Symbol</Table.Th><Table.Th>Quantity</Table.Th><Table.Th>Average cost</Table.Th><Table.Th>Cost basis</Table.Th><Table.Th ta="right">Details</Table.Th></Table.Tr></Table.Thead>
                    <Table.Tbody>
                      {holdings.map((holding) => (
                        <Table.Tr
                          key={holding.id}
                          onClick={() => setSelectedSymbol(holding.symbol)}
                          onKeyDown={(event) => {
                            if (event.key === "Enter" || event.key === " ") {
                              event.preventDefault();
                              setSelectedSymbol(holding.symbol);
                            }
                          }}
                          role="button"
                          tabIndex={0}
                          style={{ cursor: "pointer" }}
                        >
                          <Table.Td><Text fw={650}>{holding.symbol}</Text></Table.Td>
                          <Table.Td>{holding.quantity}</Table.Td>
                          <Table.Td>{formatMoney(holding.avgCost, selectedAccount.currency)}</Table.Td>
                          <Table.Td>{formatMoney(holding.avgCost * holding.quantity, selectedAccount.currency)}</Table.Td>
                          <Table.Td ta="right"><ActionIcon variant="subtle" onClick={(event) => { event.stopPropagation(); setSelectedSymbol(holding.symbol); }} aria-label={`View ${holding.symbol} holding`}><IconEye size={17} /></ActionIcon></Table.Td>
                        </Table.Tr>
                      ))}
                    </Table.Tbody>
                  </Table>
                </ScrollArea>
              ) : <EmptyState title="No holdings" detail="Filled buy orders will appear here." />}
            </Tabs.Panel>

            <Tabs.Panel value="transactions" pt="lg">
              {activity.isPending ? <Skeleton h={140} /> : activity.isError ? (
                <Text c="red.4" size="sm">Could not load account transactions.</Text>
              ) : transactions.length ? (
                <ScrollArea>
                  <Table verticalSpacing="sm" miw={680}>
                    <Table.Thead><Table.Tr><Table.Th>Type</Table.Th><Table.Th>Description</Table.Th><Table.Th>Amount</Table.Th><Table.Th>Reference</Table.Th><Table.Th ta="right">Date</Table.Th></Table.Tr></Table.Thead>
                    <Table.Tbody>
                      {transactions.map((transaction) => (
                        <Table.Tr key={transaction.id}>
                          <Table.Td><Badge variant="light">{toTitleCase(transaction.type)}</Badge></Table.Td>
                          <Table.Td>{transaction.description}</Table.Td>
                          <Table.Td>{formatMoney(transaction.amount, selectedAccount.currency)}</Table.Td>
                          <Table.Td><Text size="xs" c="dimmed">{transaction.referenceId?.slice(0, 8) ?? "—"}</Text></Table.Td>
                          <Table.Td ta="right">{formatDate(transaction.createdAt)}</Table.Td>
                        </Table.Tr>
                      ))}
                    </Table.Tbody>
                  </Table>
                </ScrollArea>
              ) : <EmptyState title="No transactions" detail="Deposits, transfers, and trades will appear here." />}
            </Tabs.Panel>
          </Tabs>
        </Paper>
      ) : null}

      <TransferDialog opened={transferOpened} accounts={accounts} initialAccountId={selectedAccount?.id} onClose={() => setTransferOpened(false)} />

      <Modal opened={Boolean(selectedSymbol)} onClose={() => setSelectedSymbol(undefined)} title={`${selectedSymbol ?? ""} holding`} centered>
        {holdingDetails.isPending ? <Skeleton h={120} /> : holdingDetails.isError ? (
          <Text c="red.4">Could not load holding details.</Text>
        ) : holdingDetails.data?.getHolding.data ? (
          <SimpleGrid cols={2}>
            <HoldingMetric label="Quantity" value={String(holdingDetails.data.getHolding.data.quantity)} />
            <HoldingMetric label="Average cost" value={formatMoney(holdingDetails.data.getHolding.data.avgCost, selectedAccount?.currency)} />
            <HoldingMetric label="Opened" value={formatDate(holdingDetails.data.getHolding.data.createdAt)} />
            <HoldingMetric label="Updated" value={formatDate(holdingDetails.data.getHolding.data.updatedAt)} />
          </SimpleGrid>
        ) : <Text c="dimmed">Holding details are unavailable.</Text>}
      </Modal>

      <Modal opened={deleteOpened} onClose={() => setDeleteOpened(false)} title="Delete paper account?" centered>
        <Stack>
          <Text>This permanently removes the simulated account, its holdings, and its transaction history.</Text>
          <Group justify="flex-end">
            <Button variant="default" onClick={() => setDeleteOpened(false)}>Keep account</Button>
            <Button color="red" loading={deleteAccount.isPending} onClick={() => selectedAccount && deleteAccount.mutate(selectedAccount.id)}>Delete paper account</Button>
          </Group>
        </Stack>
      </Modal>
    </Stack>
  );
}

function HoldingMetric({ label, value }: { label: string; value: string }) {
  return <div><Text size="xs" c="dimmed">{label}</Text><Text fw={600}>{value}</Text></div>;
}
