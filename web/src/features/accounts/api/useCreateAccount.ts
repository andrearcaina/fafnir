import { notifications } from "@mantine/notifications";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { CreatePortfolioAccountDocument } from "../../../gql/graphql";
import { graphQLClient, requireOK } from "../../../lib/api";

export interface CreateAccountValues {
  type: string;
  currency: string;
}

export function useCreateAccount(onSuccess: () => void) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (request: CreateAccountValues) => {
      const response = await graphQLClient.request(CreatePortfolioAccountDocument, { request });
      requireOK(response.createAccount.code, "Account creation");
      return response;
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["dashboard"] });
      notifications.show({
        color: "lime",
        title: "Account created",
        message: "Your new account includes $500 in simulated starting funds.",
      });
      onSuccess();
    },
    onError: (error) =>
      notifications.show({ color: "red", title: "Could not create account", message: error.message }),
  });
}
