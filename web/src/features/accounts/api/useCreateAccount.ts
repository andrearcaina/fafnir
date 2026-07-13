import { notifications } from "@mantine/notifications";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { CreatePortfolioAccountDocument } from "../../../gql/graphql";
import { graphQLClient } from "../../../lib/api";

export interface CreateAccountValues {
  type: string;
  currency: string;
}

export function useCreateAccount(onSuccess: () => void) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (request: CreateAccountValues) =>
      graphQLClient.request(CreatePortfolioAccountDocument, { request }),
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
