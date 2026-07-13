import { useEffect, useState } from "react";
import { Alert, Button, Modal, NumberInput, Select, Stack } from "@mantine/core";
import { IconInfoCircle } from "@tabler/icons-react";
import { formatMoney, toTitleCase } from "../../../lib/formatters";
import type { Account } from "../../../types/domain";
import { useDepositAccount } from "../api/useDepositAccount";

interface DepositDialogProps {
  opened: boolean;
  accounts: Account[];
  onClose: () => void;
}

export function DepositDialog({ opened, accounts, onClose }: DepositDialogProps) {
  const [accountId, setAccountId] = useState<string | null>(accounts[0]?.id ?? null);
  const [amount, setAmount] = useState<number | string>(100);
  const deposit = useDepositAccount(onClose);
  const account = accounts.find((item) => item.id === accountId);

  useEffect(() => {
    if (!accountId && accounts[0]) setAccountId(accounts[0].id);
  }, [accountId, accounts]);

  return (
    <Modal opened={opened} onClose={onClose} title="Deposit money" centered>
      <Stack gap="lg">
        <Alert color="lime" variant="light" icon={<IconInfoCircle size={17} />}>
          This is simulated money. No bank account is involved.
        </Alert>
        <Select
          label="Account"
          placeholder="Choose an account"
          value={accountId}
          onChange={setAccountId}
          data={accounts.map((item) => ({
            value: item.id,
            label: `${toTitleCase(item.type)} · ${formatMoney(item.balance)} ${item.currency}`,
          }))}
        />
        <NumberInput label="Amount" prefix="$" min={1} decimalScale={2} value={amount} onChange={setAmount} />
        <Button
          size="md"
          disabled={!account || Number(amount) <= 0}
          loading={deposit.isPending}
          onClick={() => account && deposit.mutate({ accountId: account.id, amount: Number(amount), currency: account.currency })}
        >
          Deposit {formatMoney(Number(amount) || 0)}
        </Button>
      </Stack>
    </Modal>
  );
}
