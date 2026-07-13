import { Badge, Group, Paper, Text, Title } from "@mantine/core";
import { OrdersTable } from "../../orders/components/OrdersTable";
import type { Order } from "../types";

export function OrdersSection({ orders }: { orders: Order[] }) {
  return (
    <Paper className="panel" p={{ base: "md", sm: "xl" }}>
      <Group justify="space-between" mb="xl">
        <div>
          <Title order={3}>Order history</Title>
          <Text c="dimmed" size="sm">
            All submitted paper trades
          </Text>
        </div>
        <Badge variant="light" color="gray">
          {orders.length} orders
        </Badge>
      </Group>
      <OrdersTable orders={orders} />
    </Paper>
  );
}
