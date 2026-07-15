import { useState } from "react";
import { Badge, Button, Group, Modal, Paper, SimpleGrid, Skeleton, Stack, Text, Title } from "@mantine/core";
import { OrdersTable } from "../../orders/components/OrdersTable";
import { useCancelOrder, useOrderDetails } from "../../orders/api/useOrders";
import { formatDate, formatDecimal, toTitleCase } from "../../../lib/formatters";
import type { Order } from "../types";

export function OrdersSection({ orders }: { orders: Order[] }) {
  const [selectedOrderId, setSelectedOrderId] = useState<string>();
  const details = useOrderDetails(selectedOrderId);
  const cancel = useCancelOrder(() => setSelectedOrderId(undefined));
  const order = details.data?.getOrderByOrderID.data;

  return (
    <>
      <Paper className="panel" p={{ base: "md", sm: "xl" }}>
        <Group justify="space-between" mb="xl">
          <div>
            <Title order={3}>Order history</Title>
            <Text c="dimmed" size="sm">All submitted paper trades</Text>
          </div>
          <Badge variant="light" color="gray">{orders.length} orders</Badge>
        </Group>
        <OrdersTable orders={orders} onSelect={setSelectedOrderId} />
      </Paper>

      <Modal opened={Boolean(selectedOrderId)} onClose={() => setSelectedOrderId(undefined)} title="Order details" centered>
        {details.isPending ? <Skeleton h={180} /> : details.isError ? (
          <Text c="red.4">Could not load order details.</Text>
        ) : order ? (
          <Stack gap="lg">
            <Group justify="space-between">
              <div><Text fz="xl" fw={700}>{order.symbol}</Text><Text c="dimmed" size="sm">{toTitleCase(order.side)} · {toTitleCase(order.type)}</Text></div>
              <Badge color={order.status === "FILLED" ? "lime" : order.status === "PENDING" ? "yellow" : "red"} variant="light">{toTitleCase(order.status)}</Badge>
            </Group>
            <SimpleGrid cols={2}>
              <OrderMetric label="Quantity" value={String(order.quantity)} />
              <OrderMetric label="Requested price" value={order.price ? formatDecimal(order.price) : "Market"} />
              <OrderMetric label="Filled quantity" value={String(order.filledQuantity)} />
              <OrderMetric label="Average fill" value={order.avgFillPrice ? formatDecimal(order.avgFillPrice) : "—"} />
              <OrderMetric label="Submitted" value={formatDate(order.createdAt)} />
              <OrderMetric label="Updated" value={formatDate(order.updatedAt)} />
            </SimpleGrid>
            {order.status === "PENDING" && (
              <Button color="red" variant="light" loading={cancel.isPending} onClick={() => cancel.mutate(order.id)}>
                Cancel pending order
              </Button>
            )}
          </Stack>
        ) : <Text c="dimmed">Order details are unavailable.</Text>}
      </Modal>
    </>
  );
}

function OrderMetric({ label, value }: { label: string; value: string }) {
  return <div><Text size="xs" c="dimmed">{label}</Text><Text fw={600}>{value}</Text></div>;
}
