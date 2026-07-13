import { useQuery } from "@tanstack/react-query";
import { SupportedStocksDocument } from "../../../gql/graphql";
import { graphQLClient } from "../../../lib/api";

export function useSupportedStocks() {
  return useQuery({
    queryKey: ["supported-stocks"],
    queryFn: () => graphQLClient.request(SupportedStocksDocument),
    staleTime: Number.POSITIVE_INFINITY,
    select: (data) => data.getSupportedStocks,
  });
}
