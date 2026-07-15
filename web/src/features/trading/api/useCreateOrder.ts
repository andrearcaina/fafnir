import { notifications } from "@mantine/notifications";
import { useMutation } from "@tanstack/react-query";
import { CreateOrderDocument } from "../../../gql/graphql";
import { graphQLClient, requireOK } from "../../../lib/api";

export interface CreateOrderValues {
  symbol: string;
  side: string;
  type: string;
  quantity: number;
  price?: number;
}

interface UseCreateOrderOptions {
  onSuccess: () => void;
}

export function useCreateOrder({ onSuccess }: UseCreateOrderOptions) {
  return useMutation({
    mutationFn: async (values: CreateOrderValues) => {
      const response = await graphQLClient.request(CreateOrderDocument, { request: values });
      requireOK(response.createOrder.code, "Order submission");
      return response;
    },
    onSuccess: (_data, values) => {
      notifications.show({
        color: "lime",
        title: "Order submitted",
        message: `${values.side} ${values.quantity} ${values.symbol} is now pending.`,
      });
      onSuccess();
    },
    onError: (error) =>
      notifications.show({
        color: "red",
        title: "Order rejected",
        message: error.message,
      }),
  });
}
