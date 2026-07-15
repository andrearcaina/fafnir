import { notifications } from "@mantine/notifications";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  AccountActivityDocument,
  DeletePortfolioAccountDocument,
  HoldingDetailsDocument,
  TransferFundsDocument,
  type TransferRequest,
} from "../../../gql/graphql";
import { graphQLClient, requireOK } from "../../../lib/api";

export function useAccountActivity(accountId?: string) {
  return useQuery({
    queryKey: ["account-activity", accountId],
    queryFn: async () => {
      const response = await graphQLClient.request(AccountActivityDocument, { accountId: accountId! });
      requireOK(response.getHoldings.code, "Holdings lookup");
      requireOK(response.getTransactions.code, "Transaction lookup");
      return response;
    },
    enabled: Boolean(accountId),
  });
}

export function useHoldingDetails(accountId?: string, symbol?: string) {
  return useQuery({
    queryKey: ["holding-details", accountId, symbol],
    queryFn: async () => {
      const response = await graphQLClient.request(HoldingDetailsDocument, {
        accountId: accountId!,
        symbol: symbol!,
      });
      requireOK(response.getHolding.code, "Holding lookup");
      return response;
    },
    enabled: Boolean(accountId && symbol),
  });
}

export function useTransferFunds(onSuccess: () => void) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (request: TransferRequest) => {
      const response = await graphQLClient.request(TransferFundsDocument, { request });
      requireOK(response.transfer.code, "Transfer");
      return response;
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["dashboard"] });
      await queryClient.invalidateQueries({ queryKey: ["account-activity"] });
      notifications.show({
        color: "lime",
        title: "Transfer complete",
        message: "Account balances were updated.",
      });
      onSuccess();
    },
    onError: (error) =>
      notifications.show({ color: "red", title: "Transfer failed", message: error.message }),
  });
}

export function useDeletePortfolioAccount(onSuccess: () => void) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (accountId: string) => {
      const response = await graphQLClient.request(DeletePortfolioAccountDocument, { accountId });
      if (!response.deleteAccount) throw new Error("The account could not be deleted.");
      return response;
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["dashboard"] });
      await queryClient.invalidateQueries({ queryKey: ["account-activity"] });
      notifications.show({
        color: "lime",
        title: "Account deleted",
        message: "The simulated account was removed.",
      });
      onSuccess();
    },
    onError: (error) =>
      notifications.show({ color: "red", title: "Could not delete account", message: error.message }),
  });
}
