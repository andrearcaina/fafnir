import { useEffect, useMemo, useState } from "react";
import { Alert, Button, Modal, NumberInput, Select, Stack } from "@mantine/core";
import { IconArrowsExchange } from "@tabler/icons-react";
import { formatMoney, toTitleCase } from "../../../lib/formatters";
import type { Account } from "../../../types/domain";
import { useTransferFunds } from "../api/usePortfolio";

interface TransferDialogProps {
  opened: boolean;
  accounts: Account[];
  initialAccountId?: string;
  onClose: () => void;
}

export function TransferDialog({ opened, accounts, initialAccountId, onClose }: TransferDialogProps) {
  const [fromAccountId, setFromAccountId] = useState<string | null>(initialAccountId ?? null);
  const [toAccountId, setToAccountId] = useState<string | null>(null);
  const [amount, setAmount] = useState<number | string>(100);
  const transfer = useTransferFunds(onClose);
  const fromAccount = accounts.find((account) => account.id === fromAccountId);
  const compatibleAccounts = useMemo(
    () => accounts.filter((account) => account.id !== fromAccountId && account.currency === fromAccount?.currency),
    [accounts, fromAccount, fromAccountId],
  );
  const toAccount = compatibleAccounts.find((account) => account.id === toAccountId);

  useEffect(() => {
    if (opened) setFromAccountId(initialAccountId ?? accounts[0]?.id ?? null);
  }, [accounts, initialAccountId, opened]);

  useEffect(() => {
    if (!compatibleAccounts.some((account) => account.id === toAccountId)) {
      setToAccountId(compatibleAccounts[0]?.id ?? null);
    }
  }, [compatibleAccounts, toAccountId]);

  const accountOptions = (items: Account[]) =>
    items.map((account) => ({
      value: account.id,
      label: `${toTitleCase(account.type)} · ${formatMoney(account.balance, account.currency)}`,
    }));

  return (
    <Modal opened={opened} onClose={onClose} title="Transfer between accounts" centered>
      <Stack gap="lg">
        <Alert color="blue" variant="light" icon={<IconArrowsExchange size={17} />}>
          Transfers require two accounts using the same currency.
        </Alert>
        <Select label="From" value={fromAccountId} onChange={setFromAccountId} data={accountOptions(accounts)} />
        <Select
          label="To"
          value={toAccountId}
          onChange={setToAccountId}
          data={accountOptions(compatibleAccounts)}
          placeholder={fromAccount ? `Choose another ${fromAccount.currency} account` : "Choose an account"}
          disabled={!compatibleAccounts.length}
        />
        <NumberInput
          label="Amount"
          prefix="$"
          min={0.01}
          max={fromAccount?.balance}
          decimalScale={2}
          value={amount}
          onChange={setAmount}
        />
        <Button
          disabled={!fromAccount || !toAccount || Number(amount) <= 0 || Number(amount) > fromAccount.balance}
          loading={transfer.isPending}
          onClick={() =>
            fromAccount &&
            toAccount &&
            transfer.mutate({
              fromAccountId: fromAccount.id,
              toAccountId: toAccount.id,
              amount: Number(amount),
              currency: fromAccount.currency,
            })
          }
        >
          Transfer {formatMoney(Number(amount) || 0, fromAccount?.currency)}
        </Button>
      </Stack>
    </Modal>
  );
}
