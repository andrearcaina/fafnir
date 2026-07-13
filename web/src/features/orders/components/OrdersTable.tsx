import { Badge, ScrollArea, Table, Text } from "@mantine/core";
import { EmptyState } from "../../../components/feedback/EmptyState";
import { formatDate, formatMoney, toTitleCase } from "../../../lib/formatters";
import type { Order } from "../../../types/domain";

export function OrdersTable({ orders }: { orders: Order[] }) {
  if (!orders.length) {
    return (
      <EmptyState
        title="No orders yet"
        detail="Your first simulated trade will show up here."
      />
    );
  }

  return (
    <ScrollArea>
      <Table verticalSpacing="md" horizontalSpacing="md" miw={700} className="orders-table">
        <Table.Thead>
          <Table.Tr>
            <Table.Th>Symbol</Table.Th>
            <Table.Th>Side</Table.Th>
            <Table.Th>Type</Table.Th>
            <Table.Th>Quantity</Table.Th>
            <Table.Th>Price</Table.Th>
            <Table.Th>Status</Table.Th>
            <Table.Th ta="right">Submitted</Table.Th>
          </Table.Tr>
        </Table.Thead>
        <Table.Tbody>
          {orders.map((order) => (
            <Table.Tr key={order.id}>
              <Table.Td>
                <Text fw={650} size="sm">
                  {order.symbol}
                </Text>
              </Table.Td>
              <Table.Td>
                <Text c={order.side === "BUY" ? "lime.5" : "red.5"} fw={600} size="sm">
                  {toTitleCase(order.side)}
                </Text>
              </Table.Td>
              <Table.Td>{toTitleCase(order.type)}</Table.Td>
              <Table.Td>{order.quantity}</Table.Td>
              <Table.Td>{order.price ? formatMoney(order.price) : "Market"}</Table.Td>
              <Table.Td>
                <Badge color={getStatusColor(order.status)} variant="light" size="sm">
                  {toTitleCase(order.status)}
                </Badge>
              </Table.Td>
              <Table.Td ta="right" c="dimmed">
                {formatDate(order.createdAt)}
              </Table.Td>
            </Table.Tr>
          ))}
        </Table.Tbody>
      </Table>
    </ScrollArea>
  );
}

function getStatusColor(status: string) {
  if (status === "FILLED") return "lime";
  if (status === "CANCELLED" || status === "REJECTED") return "red";
  return "yellow";
}
