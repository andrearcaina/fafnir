import { notifications } from "@mantine/notifications";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { CancelOrderDocument, OrderDetailsDocument } from "../../../gql/graphql";
import { graphQLClient, requireOK } from "../../../lib/api";

export function useOrderDetails(orderId?: string) {
  return useQuery({
    queryKey: ["order-details", orderId],
    queryFn: async () => {
      const response = await graphQLClient.request(OrderDetailsDocument, { orderId: orderId! });
      requireOK(response.getOrderByOrderID.code, "Order lookup");
      return response;
    },
    enabled: Boolean(orderId),
  });
}

export function useCancelOrder(onSuccess: () => void) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (orderId: string) => {
      const response = await graphQLClient.request(CancelOrderDocument, { orderId });
      requireOK(response.cancelOrder.code, "Cancellation");
      return response;
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["dashboard"] });
      await queryClient.invalidateQueries({ queryKey: ["order-details"] });
      notifications.show({
        color: "lime",
        title: "Order cancelled",
        message: "The pending order was cancelled.",
      });
      onSuccess();
    },
    onError: (error) =>
      notifications.show({ color: "red", title: "Could not cancel order", message: error.message }),
  });
}
