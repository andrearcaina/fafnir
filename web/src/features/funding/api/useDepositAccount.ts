import { notifications } from "@mantine/notifications";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { DepositFundsDocument } from "../../../gql/graphql";
import { graphQLClient, requireOK } from "../../../lib/api";

export interface DepositValues {
  accountId: string;
  amount: number;
  currency: string;
}

export function useDepositAccount(onSuccess: () => void) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (request: DepositValues) => {
      const response = await graphQLClient.request(DepositFundsDocument, { request });
      requireOK(response.deposit.code, "Deposit");
      return response;
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["dashboard"] });
      await queryClient.invalidateQueries({ queryKey: ["account-activity"] });
      notifications.show({
        color: "lime",
        title: "Money deposited",
        message: "Your simulated account balance has been updated.",
      });
      onSuccess();
    },
    onError: (error) =>
      notifications.show({
        color: "red",
        title: "Deposit failed",
        message: error.message,
      }),
  });
}
