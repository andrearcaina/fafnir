import { notifications } from "@mantine/notifications";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import {
  AddStockToWatchlistDocument,
  RemoveStockFromWatchlistDocument,
} from "../../../gql/graphql";
import { graphQLClient, requireOK } from "../../../lib/api";

export function useWatchlist(symbol: string, isSaved: boolean) {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: async () => {
      if (isSaved) {
        const response = await graphQLClient.request(RemoveStockFromWatchlistDocument, {
          request: { symbol },
        });
        requireOK(response.removeFromWatchlist.code, "Watchlist removal");
        return;
      }

      const response = await graphQLClient.request(AddStockToWatchlistDocument, {
        request: { symbol },
      });
      requireOK(response.addToWatchlist.code, "Watchlist update");
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["dashboard"] });
      notifications.show({
        color: "lime",
        title: isSaved ? "Removed from watchlist" : "Added to watchlist",
        message: `${symbol} ${isSaved ? "was removed from" : "is now in"} your watchlist.`,
      });
    },
    onError: (error) =>
      notifications.show({
        color: "red",
        title: "Could not update watchlist",
        message: error.message,
      }),
  });

  return {
    toggle: () => mutation.mutate(),
    isPending: mutation.isPending,
  };
}
