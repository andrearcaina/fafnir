import { useState } from "react";
import { Alert, Button, Modal, Select, Stack } from "@mantine/core";
import { IconSparkles } from "@tabler/icons-react";
import { useCreateAccount } from "../api/useCreateAccount";

interface CreateAccountDialogProps {
  opened: boolean;
  onClose: () => void;
}

export function CreateAccountDialog({ opened, onClose }: CreateAccountDialogProps) {
  const [type, setType] = useState<string | null>("INVESTMENT");
  const [currency, setCurrency] = useState<string | null>("USD");
  const createAccount = useCreateAccount(onClose);

  return (
    <Modal opened={opened} onClose={onClose} title="Open an account" centered>
      <Stack gap="lg">
        <Alert color="lime" variant="light" icon={<IconSparkles size={17} />}>
          New accounts start with $500 in simulated funds.
        </Alert>
        <Select
          label="Account type"
          value={type}
          onChange={setType}
          data={[
            { value: "INVESTMENT", label: "Investment" },
            { value: "SAVINGS", label: "Savings" },
            { value: "CHEQUING", label: "Chequing" },
          ]}
        />
        <Select
          label="Currency"
          value={currency}
          onChange={setCurrency}
          data={["USD", "CAD"]}
        />
        <Button
          size="md"
          disabled={!type || !currency}
          loading={createAccount.isPending}
          onClick={() => type && currency && createAccount.mutate({ type, currency })}
        >
          Open account
        </Button>
      </Stack>
    </Modal>
  );
}
