import { useState } from "react";
import {
  Button,
  Divider,
  Group,
  NumberInput,
  Paper,
  SegmentedControl,
  Select,
  Stack,
  Text,
  TextInput,
} from "@mantine/core";
import { formatMoney, toTitleCase } from "../../../lib/formatters";
import { useCreateOrder } from "../api/useCreateOrder";

interface OrderTicketProps {
  defaultSymbol: string;
  currency?: string;
  onComplete: () => void;
}

export function OrderTicket({ defaultSymbol, currency, onComplete }: OrderTicketProps) {
  const [side, setSide] = useState("BUY");
  const [type, setType] = useState("MARKET");
  const [quantity, setQuantity] = useState<number | string>(1);
  const [price, setPrice] = useState<number | string>("");
  const createOrder = useCreateOrder({ onSuccess: onComplete });

  const canSubmit =
    defaultSymbol.length > 0 &&
    Number(quantity) > 0 &&
    (type === "MARKET" || Number(price) > 0);

  const submit = () => {
    createOrder.mutate({
      symbol: defaultSymbol,
      side,
      type,
      quantity: Number(quantity),
      ...(type !== "MARKET" && price ? { price: Number(price) } : {}),
    });
  };

  return (
    <Stack gap="lg">
      <Paper className="ticket-summary" p="lg">
        <Text c="dimmed" size="xs">
          SIMULATED ORDER
        </Text>
        <Text fw={650} mt={4}>
          No real money will be used.
        </Text>
      </Paper>
      <SegmentedControl
        fullWidth
        value={side}
        onChange={setSide}
        color={side === "BUY" ? "lime" : "red"}
        data={["BUY", "SELL"]}
      />
      <TextInput
        label="Symbol"
        value={defaultSymbol}
        readOnly
      />
      <Select
        label="Order type"
        value={type}
        onChange={(value) => setType(value ?? "MARKET")}
        data={[
          { value: "MARKET", label: "Market" },
          { value: "LIMIT", label: "Limit" },
        ]}
      />
      <NumberInput
        label="Quantity"
        min={0.0001}
        decimalScale={4}
        value={quantity}
        onChange={setQuantity}
      />
      {type !== "MARKET" && (
        <NumberInput
          label="Limit price"
          prefix={currency ? `${currency} ` : undefined}
          min={0.01}
          decimalScale={2}
          value={price}
          onChange={setPrice}
        />
      )}
      <Divider />
      <Group justify="space-between">
        <Text c="dimmed">Estimated value</Text>
        <Text fw={650}>
          {price && quantity
            ? formatMoney(Number(price) * Number(quantity), currency)
            : "Calculated at market"}
        </Text>
      </Group>
      <Button
        size="md"
        color={side === "BUY" ? "lime" : "red"}
        disabled={!canSubmit}
        loading={createOrder.isPending}
        onClick={submit}
      >
        {toTitleCase(side)} {defaultSymbol || "stock"}
      </Button>
    </Stack>
  );
}
