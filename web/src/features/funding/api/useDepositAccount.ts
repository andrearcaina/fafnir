import { notifications } from "@mantine/notifications";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { DepositFundsDocument } from "../../../gql/graphql";
import { graphQLClient } from "../../../lib/api";

export interface DepositValues {
  accountId: string;
  amount: number;
  currency: string;
}

export function useDepositAccount(onSuccess: () => void) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (request: DepositValues) =>
      graphQLClient.request(DepositFundsDocument, { request }),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["dashboard"] });
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
